package cli

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ViitoJooj/devx/internal/cli/config"
	"github.com/ViitoJooj/devx/internal/cli/utils"
	"github.com/ViitoJooj/devx/internal/templates"
	"github.com/ViitoJooj/devx/pkg/errorx"
	"github.com/spf13/cobra"
)

var serviceRmCmd = &cobra.Command{
	Use:       "rm <service...>",
	Aliases:   []string{"remove"},
	Short:     "Remove services from the project (migrations are kept)",
	Args:      cobra.MinimumNArgs(1),
	ValidArgs: serviceOrder,
	Run:       serviceRm,
}

func init() {
	serviceRmCmd.Flags().BoolVarP(&serviceRmYes, "yes", "y", false, "skip the confirmation")
}

var serviceRmYes bool

func serviceRm(cmd *cobra.Command, args []string) {

	project := loadProject()

	for _, name := range args {
		svc, ok := services[name]
		if !ok {
			errorx.UnknownService(name)
		}

		if !project.Exists(svc.mainPath()...) {
			errorx.ServiceNeedsApp(name, svc.app())
		}

		if !project.Exists(svc.detectPath()...) {
			errorx.ServiceNotAdded(name)
		}
	}

	for _, name := range args {
		if usedBy := serviceUsedBy(project, services[name], args); usedBy != "" {
			errorx.ServiceInUse(name, usedBy)
		}
	}

	confirm("Remove "+strings.Join(args, ", ")+" from "+project.Name+"?", serviceRmYes)

	for _, name := range args {
		if err := removeService(project, services[name]); err != nil {
			errorx.CannotGenerate(err)
		}
		fmt.Println(name + " removed")
	}

	if err := utils.RunGo(project.Root, "mod", "tidy"); err != nil {
		errorx.CannotRunGo(err)
	}
}

// serviceUsedBy returns who still depends on svc, ignoring services removed in the same command.
func serviceUsedBy(project *config.Project, svc service, removing []string) string {

	var users []string

	for _, name := range serviceOrder {
		other := services[name]
		if contains(removing, name) || !project.Exists(other.detectPath()...) {
			continue
		}
		if contains(other.Requires, svc.Name) {
			users = append(users, name)
		}
	}

	if svc.Name == "postgres" {
		for _, name := range cruds(project) {
			if crudDatabase(project, name) == "postgres" {
				users = append(users, "crud "+name)
			}
		}
	}

	return strings.Join(users, ", ")
}

func removeService(project *config.Project, svc service) error {

	data := projectData(project)

	if err := removeServiceFiles(project, svc); err != nil {
		return err
	}

	if svc.Env != "" {
		env, err := utils.Render("env", svc.Env, data)
		if err != nil {
			return err
		}

		for _, file := range []string{".env", ".env.example"} {
			if err := utils.RemoveEnvBlock(project.Path(file), env); err != nil {
				return err
			}
		}

		dotenv := project.Path("pkg", "dotenv", "dotenv.go")
		for _, snippet := range []string{svc.CfgValidate, svc.CfgValue, svc.CfgType, svc.CfgField} {
			if err := utils.RemoveSnippet(dotenv, snippet); err != nil {
				return err
			}
		}
	}

	main := project.Path(svc.mainPath()...)
	for _, snippet := range []string{svc.Conn, svc.Middleware} {
		if err := utils.RemoveSnippet(main, snippet); err != nil {
			return err
		}
	}

	for _, imp := range svc.Imports {
		line, err := utils.Render("import", imp, data)
		if err != nil {
			return err
		}

		if err := utils.RemoveSnippet(main, line); err != nil {
			return err
		}
	}

	if svc.app() == "api" {
		if err := removeCompose(project, svc); err != nil {
			return err
		}
	}

	return syncMakefile(project)
}

func removeServiceFiles(project *config.Project, svc service) error {

	source := "services/" + svc.Name

	return fs.WalkDir(templates.FS, source, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		rel := strings.TrimSuffix(strings.TrimPrefix(path, source+"/"), ".tmpl")
		if strings.HasPrefix(rel, "migrations/") {
			return nil
		}

		dest := project.Path(rel)
		if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
			return err
		}

		utils.RemoveEmptyDirs(project.Root, filepath.Dir(dest))
		return nil
	})
}

func removeCompose(project *config.Project, svc service) error {

	path := composePath(project)
	if !project.Exists("infra", "compose", "docker-compose.yaml") {
		return nil
	}

	if svc.Compose != "" {
		if err := utils.YamlRemove(path, []string{"services", svc.composeKey()}); err != nil {
			return err
		}

		if svc.Volume != "" {
			if err := utils.YamlRemove(path, []string{"volumes", strings.TrimSuffix(svc.Volume, ":")}); err != nil {
				return err
			}
		}

		if !utils.YamlHasChildren(path, []string{"volumes"}) {
			if err := utils.YamlRemove(path, []string{"volumes"}); err != nil {
				return err
			}
		}
	}

	for line := range strings.SplitSeq(svc.DockerEnv, "\n") {
		key, _, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		if err := utils.YamlRemove(path, []string{"services", "api", "environment", key}); err != nil {
			return err
		}
	}

	if err := updateDockerDepends(project); err != nil {
		return err
	}

	if utils.YamlHasChildren(path, []string{"services"}) {
		return nil
	}

	if err := os.Remove(path); err != nil {
		return err
	}

	utils.RemoveEmptyDirs(project.Root, filepath.Dir(path))
	return nil
}

func contains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}
