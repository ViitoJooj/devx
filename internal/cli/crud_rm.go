package cli

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/ViitoJooj/devx/internal/cli/config"
	"github.com/ViitoJooj/devx/internal/cli/utils"
	"github.com/ViitoJooj/devx/internal/templates"
	"github.com/ViitoJooj/devx/pkg/errorx"
	"github.com/spf13/cobra"
)

var entityType = regexp.MustCompile(`type (\w+) struct`)

var crudRmYes bool

var crudRmCmd = &cobra.Command{
	Use:     "rm <name>",
	Aliases: []string{"remove"},
	Short:   "Remove a container and its hacks, routes and openapi (migrations are kept)",
	Args:    cobra.ExactArgs(1),
	Run:     crudRm,
}

func init() {
	crudRmCmd.Flags().BoolVarP(&crudRmYes, "yes", "y", false, "skip the confirmation")
}

func crudRm(cmd *cobra.Command, args []string) {

	project := loadAPIProject()

	if isBaseServer(project, args[0]) {
		removeServer(project)
		return
	}

	names := resolveCrud(project, args[0])
	names.Pascal = entityName(project, names)
	db := crudDatabase(project, names.FolderPlural)

	confirm("Remove crud "+names.Plural+" from "+project.Name+"?", crudRmYes)

	for _, dir := range [][]string{{"internal", "containers", names.FolderPlural}, {"hacks", "http", names.KebabPlural}} {
		if err := os.RemoveAll(project.Path(dir...)); err != nil {
			errorx.CannotGenerate(err)
		}
	}

	if len(cruds(project)) == 0 {
		if err := os.Remove(project.Path("pkg", "utils", "http_utils.go")); err != nil && !os.IsNotExist(err) {
			errorx.CannotGenerate(err)
		}
	}

	main := project.Path("cmd", "api", "main.go")

	if err := utils.RemoveSnippet(main, containerInit(names, db)); err != nil {
		errorx.CannotGenerate(err)
	}

	if err := utils.RemoveSnippet(main, containerImport(project, names)); err != nil {
		errorx.CannotGenerate(err)
	}

	if err := removeOpenAPI(project, names); err != nil {
		errorx.CannotGenerate(err)
	}

	fmt.Println("removed internal/containers/" + names.FolderPlural + " and hacks/http/" + names.KebabPlural)
	fmt.Println("migrations were kept: to drop the table, create a new migration with DROP TABLE " + names.Plural)
}

func entityName(project *config.Project, names utils.Names) string {

	data, err := os.ReadFile(project.Path("internal", "containers", names.FolderPlural, "entities", "entitie.go"))
	if err != nil {
		return names.Pascal
	}

	match := entityType.FindStringSubmatch(string(data))
	if match == nil {
		return names.Pascal
	}

	return match[1]
}

// isBaseServer tells if input names the api container (server) and not a crud called servers.
func isBaseServer(project *config.Project, input string) bool {

	input = strings.ToLower(input)
	if input != "server" && input != "servers" {
		return false
	}

	return project.Exists("internal", "containers", "server") && !slices.Contains(cruds(project), "servers")
}

func removeServer(project *config.Project) {

	confirm("Remove the api container server (/v1/ping, /v1/openapi.yaml and /docs) from "+project.Name+"?", crudRmYes)

	for _, dir := range [][]string{{"internal", "containers", "server"}, {"hacks", "http", "server"}} {
		if err := os.RemoveAll(project.Path(dir...)); err != nil {
			errorx.CannotGenerate(err)
		}
	}

	main := project.Path("cmd", "api", "main.go")

	if err := utils.RemoveSnippet(main, "if err := server.Init(ctx, api); err != nil {\n\terrorx.Fatal(err)\n}"); err != nil {
		errorx.CannotGenerate(err)
	}

	if err := utils.RemoveSnippet(main, `"`+project.Module+`/internal/containers/server"`); err != nil {
		errorx.CannotGenerate(err)
	}

	contract, err := serverContract()
	if err != nil {
		errorx.CannotGenerate(err)
	}

	if err := utils.RemoveSnippet(project.Path("internal", "contracts", "repository.go"), contract); err != nil {
		errorx.CannotGenerate(err)
	}

	fmt.Println("removed internal/containers/server and hacks/http/server (/v1/ping, /v1/openapi.yaml and /docs)")
}

// serverContract reads IServerRepository from the contracts template, so the removed text always matches the generated one.
func serverContract() (string, error) {

	data, err := templates.FS.ReadFile("api/internal/contracts/repository.go.tmpl")
	if err != nil {
		return "", err
	}

	_, rest, ok := strings.Cut(string(data), "type IServerRepository interface {")
	if !ok {
		return "", nil
	}

	body, _, _ := strings.Cut(rest, "\n}")
	return "type IServerRepository interface {" + body + "\n}", nil
}
