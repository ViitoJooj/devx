package cli

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ViitoJooj/devx/internal/cli/config"
	"github.com/ViitoJooj/devx/internal/cli/utils"
	"github.com/ViitoJooj/devx/pkg/errorx"
	"github.com/spf13/cobra"
)

var envValue = regexp.MustCompile(`(?m)^([A-Z0-9_]+)=.*$`)

var serviceAddCmd = &cobra.Command{
	Use:       "add <service...>",
	Short:     "Add services (" + strings.Join(serviceOrder, ", ") + ") to the project",
	Args:      cobra.MinimumNArgs(1),
	ValidArgs: serviceOrder,
	Run:       serviceAdd,
}

func serviceAdd(cmd *cobra.Command, args []string) {

	project := loadProject()

	for _, name := range args {
		svc, ok := services[name]
		if !ok {
			errorx.UnknownService(name)
		}

		if !project.Exists(svc.mainPath()...) {
			errorx.ServiceNeedsApp(name, svc.app())
		}
	}

	for _, name := range args {
		if err := addService(project, services[name]); err != nil {
			errorx.CannotGenerate(err)
		}
	}

	if err := utils.RunGo(project.Root, "mod", "tidy"); err != nil {
		errorx.CannotRunGo(err)
	}
}

func addService(project *config.Project, svc service) error {

	if project.Exists(svc.detectPath()...) {
		fmt.Println(svc.Name + " already added, skipping")
		return nil
	}

	for _, required := range svc.Requires {
		if project.Exists(services[required].detectPath()...) {
			continue
		}

		if err := addService(project, services[required]); err != nil {
			return err
		}
	}

	data := projectData(project)

	_, err := utils.RenderDir(utils.RenderInput{
		Source:       "services/" + svc.Name,
		Destination:  project.Root,
		Data:         data,
		SkipExisting: true,
	})
	if err != nil {
		return err
	}

	steps := []func() error{
		func() error { return addEnv(project, svc, data) },
		func() error { return addDotenv(project, svc) },
		func() error { return addMain(project, svc, data) },
		func() error { return addCompose(project, svc, data) },
		func() error { return wireDocker(project, svc) },
		func() error { return syncMakefile(project) },
	}

	for _, step := range steps {
		if err := step(); err != nil {
			return err
		}
	}

	for _, dep := range svc.Deps {
		if err := utils.RunGo(project.Root, "get", dep); err != nil {
			return err
		}
	}

	fmt.Println(svc.Name + " added")
	return nil
}

func addEnv(project *config.Project, svc service, data templateData) error {

	if svc.Env == "" {
		return nil
	}

	env, err := utils.Render("env", svc.Env, data)
	if err != nil {
		return err
	}

	if err := utils.AppendFile(project.Path(".env"), env); err != nil {
		return err
	}

	return utils.AppendFile(project.Path(".env.example"), envValue.ReplaceAllString(env, "$1="))
}

func addDotenv(project *config.Project, svc service) error {

	if svc.CfgField == "" {
		return nil
	}

	path := project.Path("pkg", "dotenv", "dotenv.go")

	injections := []struct {
		anchor  utils.Anchor
		snippet string
	}{
		{cfgFieldAnchor, svc.CfgField},
		{cfgTypeAnchor, svc.CfgType + "\n"},
		{cfgValueAnchor, svc.CfgValue},
		{cfgValidateAnchor, svc.CfgValidate + "\n"},
	}

	for _, injection := range injections {
		if err := utils.InsertGo(path, injection.anchor, injection.snippet); err != nil {
			return err
		}
	}

	return nil
}

func addMain(project *config.Project, svc service, data templateData) error {

	path := project.Path(svc.mainPath()...)

	for _, imp := range svc.Imports {
		line, err := utils.Render("import", imp, data)
		if err != nil {
			return err
		}

		if err := addImport(path, line); err != nil {
			return err
		}
	}

	if svc.Middleware != "" {
		if err := utils.InsertGo(path, middlewareAnchor, svc.Middleware+"\n"); err != nil {
			return err
		}
	}

	if svc.Conn == "" {
		return nil
	}

	if svc.app() == "cli" {
		return utils.InsertGo(path, beforeStmtAnchor("main", "commands.Execute()"), svc.Conn)
	}

	return utils.InsertGo(path, connAnchor, svc.Conn)
}

func composePath(project *config.Project) string {
	return project.Path("infra", "compose", "docker-compose.yaml")
}

func addCompose(project *config.Project, svc service, data templateData) error {

	if svc.Compose == "" {
		return nil
	}

	path := composePath(project)

	if !project.Exists("infra", "compose", "docker-compose.yaml") {
		if err := utils.WriteNewFile(path, "name: "+project.Name+"\n\nservices:\n"); err != nil {
			return err
		}
	}

	compose, err := utils.Render("compose", svc.Compose, data)
	if err != nil {
		return err
	}

	if err := utils.YamlAppend(path, []string{"services"}, compose, true); err != nil {
		return err
	}

	if svc.Volume == "" {
		return nil
	}

	return utils.YamlAppend(path, []string{"volumes"}, svc.Volume, false)
}

func projectData(project *config.Project) templateData {
	return templateData{
		Module:    project.Module,
		Name:      project.Name,
		Port:      project.Port(),
		GoVersion: project.GoVersion(),
	}
}

func wireDocker(project *config.Project, svc service) error {

	if !project.Exists(dockerService.detectPath()...) {
		return nil
	}

	if svc.Name == dockerService.Name {
		for _, name := range serviceOrder {
			other := services[name]
			if other.Name != dockerService.Name && project.Exists(other.detectPath()...) {
				if err := wireDocker(project, other); err != nil {
					return err
				}
			}
		}
		return updateDockerDepends(project)
	}

	if svc.DockerEnv != "" {
		if err := utils.YamlAppend(composePath(project), []string{"services", "api", "environment"}, svc.DockerEnv, false); err != nil {
			return err
		}
	}

	return updateDockerDepends(project)
}

func updateDockerDepends(project *config.Project) error {

	if !project.Exists(dockerService.detectPath()...) {
		return nil
	}

	path := composePath(project)

	if err := utils.YamlRemove(path, []string{"services", "api", "depends_on"}); err != nil {
		return err
	}

	var depends []string
	for _, name := range serviceOrder {
		svc := services[name]
		if svc.Healthy && project.Exists(svc.detectPath()...) {
			depends = append(depends, name+":\n  condition: service_healthy")
		}
	}

	if len(depends) == 0 {
		return nil
	}

	return utils.YamlAppend(path, []string{"services", "api", "depends_on"}, strings.Join(depends, "\n"), false)
}
