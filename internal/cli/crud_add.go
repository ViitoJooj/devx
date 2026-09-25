package cli

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/ViitoJooj/devx/internal/cli/config"
	"github.com/ViitoJooj/devx/internal/cli/utils"
	"github.com/ViitoJooj/devx/pkg/errorx"
	"github.com/spf13/cobra"
)

var migrationNumber = regexp.MustCompile(`^(\d+)_`)

var (
	crudNoMigrations bool
	crudAddDB        string
)

var crudAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Create a container with create, list, query, get, update, upgrade and delete",
	Args:  cobra.ExactArgs(1),
	Run:   crudAdd,
}

func init() {
	crudAddCmd.Flags().BoolVar(&crudNoMigrations, "no-migrations", false, "do not create the migration (table already exists)")
	crudAddCmd.Flags().StringVar(&crudAddDB, "db", "", "repository: postgres or memory (asked when the project has no database)")
}

func crudAdd(cmd *cobra.Command, args []string) {

	project := loadAPIProject()

	names, err := utils.NewEntityNames(args[0])
	if err != nil {
		errorx.InvalidName(args[0])
	}

	if project.Exists("internal", "containers", names.FolderPlural) {
		errorx.ContainerAlreadyExists(names.FolderPlural)
	}

	db := chooseDatabase(project, crudAddDB)

	if db == "postgres" && !project.Exists("pkg", "postgres") {
		fmt.Println("adding postgres")
		if err := addService(project, services["postgres"]); err != nil {
			errorx.CannotGenerate(err)
		}
	}

	var exclude []string
	if crudNoMigrations || db != "postgres" {
		exclude = append(exclude, "migrations/")
	}

	migration, err := nextMigration(project, names)
	if err != nil {
		errorx.CannotGenerate(err)
	}

	created, err := utils.RenderDir(utils.RenderInput{
		Source:      "crud",
		Destination: project.Root,
		Data: templateData{
			Module:    project.Module,
			Name:      project.Name,
			Migration: migration,
			DB:        db,
			Names:     names,
		},
		Placeholders: map[string]string{
			"__plural__":    names.FolderPlural,
			"__kebab__":     names.KebabPlural,
			"__migration__": migration,
		},
		Exclude:      exclude,
		SkipExisting: true,
	})
	if err != nil {
		errorx.CannotGenerate(err)
	}

	if err := renderCrudDB(project, names, db, false); err != nil {
		errorx.CannotGenerate(err)
	}

	if err := addContainerToMain(project, names, db); err != nil {
		errorx.CannotGenerate(err)
	}

	if err := addOpenAPI(project, names); err != nil {
		errorx.CannotGenerate(err)
	}

	if err := utils.RunGo(project.Root, "mod", "tidy"); err != nil {
		errorx.CannotRunGo(err)
	}

	for _, file := range created {
		fmt.Println("created " + file)
	}
}

func addContainerToMain(project *config.Project, names utils.Names, db string) error {

	path := project.Path("cmd", "api", "main.go")

	if err := addImport(path, containerImport(project, names)); err != nil {
		return err
	}

	g, err := utils.LoadGo(path)
	if err != nil {
		return err
	}

	snippet := containerInit(names, db)
	if g.LastStmt("main", isContainerInit) == nil {
		snippet += "\n\n"
	}

	return utils.InsertGo(path, containerAnchor, snippet)
}

func containerImport(project *config.Project, names utils.Names) string {
	return names.Plural + ` "` + project.Module + "/internal/containers/" + names.FolderPlural + `"`
}

func containerInit(names utils.Names, db string) string {
	return `if err := ` + containerCall(names, db) + `; err != nil {
	errorx.Fatal(err)
}`
}

func nextMigration(project *config.Project, names utils.Names) (string, error) {

	number, err := nextMigrationNumber(project)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%04d_%s_table", number, names.Pascal), nil
}

func nextMigrationNumber(project *config.Project) (int, error) {

	entries, err := os.ReadDir(project.Path("migrations"))
	if err != nil && !os.IsNotExist(err) {
		return 0, err
	}

	last := 0
	for _, entry := range entries {
		match := migrationNumber.FindStringSubmatch(entry.Name())
		if match == nil {
			continue
		}

		number, _ := strconv.Atoi(match[1])
		last = max(last, number)
	}

	return last + 1, nil
}
