package cli

import (
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strings"

	"github.com/ViitoJooj/devx/internal/cli/config"
	"github.com/ViitoJooj/devx/internal/cli/utils"
	"github.com/ViitoJooj/devx/internal/templates"
	"github.com/ViitoJooj/devx/pkg/errorx"
	"github.com/spf13/cobra"
)

var crudDBs = []string{"postgres", "memory"}

var crudDBYes bool

var crudDBCmd = &cobra.Command{
	Use:       "db <name> <postgres|memory>",
	Short:     "Switch the repository of a crud between postgres and memory",
	Args:      cobra.ExactArgs(2),
	ValidArgs: crudDBs,
	Run:       crudDB,
}

func init() {
	crudDBCmd.Flags().BoolVarP(&crudDBYes, "yes", "y", false, "skip the confirmation")
	crudDBCmd.Flags().BoolVar(&crudNoMigrations, "no-migrations", false, "do not create the migration when switching to postgres")
}

func crudDB(cmd *cobra.Command, args []string) {

	project := loadAPIProject()

	names := resolveCrud(project, args[0])
	names.Pascal = entityName(project, names)

	db := args[1]
	if !slices.Contains(crudDBs, db) {
		errorx.InvalidDB(db)
	}

	current := crudDatabase(project, names.FolderPlural)
	if current == db {
		fmt.Println(names.Plural + " already uses " + db)
		return
	}

	confirm("Replace repositories/repository.go and initializr.go of "+names.Plural+" with "+db+" (changes made by hand in them are lost)?", crudDBYes)

	if err := prepareDatabase(project, names, db); err != nil {
		errorx.CannotGenerate(err)
	}

	if err := renderCrudDB(project, names, db, true); err != nil {
		errorx.CannotGenerate(err)
	}

	main := project.Path("cmd", "api", "main.go")
	data, err := os.ReadFile(main)
	if err != nil {
		errorx.CannotGenerate(err)
	}

	content := strings.Replace(string(data), containerCall(names, current), containerCall(names, db), 1)
	if err := os.WriteFile(main, []byte(content), 0o644); err != nil {
		errorx.CannotGenerate(err)
	}

	fmt.Println(names.Plural + " now uses " + db)
}

// chooseDatabase: --db flag, postgres when the project has it, a question in the terminal or an error in scripts.
func chooseDatabase(project *config.Project, flag string) string {

	if flag != "" {
		if !slices.Contains(crudDBs, flag) {
			errorx.InvalidDB(flag)
		}
		return flag
	}

	if project.Exists("pkg", "postgres") {
		return "postgres"
	}

	switch choose("This project has no database. What should the crud use?", []string{
		"add postgres",
		"memory repository (data is lost on restart, no migration)",
		"cancel",
	}) {
	case 0:
		return "postgres"
	case 1:
		return "memory"
	default:
		fmt.Println("aborted")
		os.Exit(0)
		return ""
	}
}

// prepareDatabase adds the postgres service and the migration when the crud goes to postgres.
func prepareDatabase(project *config.Project, names utils.Names, db string) error {

	if db != "postgres" {
		return nil
	}

	if !project.Exists("pkg", "postgres") {
		if err := addService(project, services["postgres"]); err != nil {
			return err
		}
	}

	if crudNoMigrations || hasMigration(project, names) {
		return nil
	}

	return renderMigration(project, names)
}

func renderCrudDB(project *config.Project, names utils.Names, db string, overwrite bool) error {

	data := templateData{Module: project.Module, Name: project.Name, DB: db, Names: names}

	files := []struct {
		template string
		dest     []string
	}{
		{"crud_db/" + db + "/repository.go.tmpl", []string{"internal", "containers", names.FolderPlural, "repositories", "repository.go"}},
		{"crud/internal/containers/__plural__/initializr.go.tmpl", []string{"internal", "containers", names.FolderPlural, "initializr.go"}},
	}

	for _, file := range files {
		if !overwrite && file.dest[len(file.dest)-1] == "initializr.go" {
			continue
		}

		content, err := fs.ReadFile(templates.FS, file.template)
		if err != nil {
			return err
		}

		output, err := utils.Render(file.template, string(content), data)
		if err != nil {
			return err
		}

		if err := os.MkdirAll(project.Path(file.dest[:len(file.dest)-1]...), 0o755); err != nil {
			return err
		}

		if err := os.WriteFile(project.Path(file.dest...), []byte(output), 0o644); err != nil {
			return err
		}
	}

	return nil
}

func renderMigration(project *config.Project, names utils.Names) error {

	migration, err := nextMigration(project, names)
	if err != nil {
		return err
	}

	data := templateData{Module: project.Module, Name: project.Name, Migration: migration, Names: names}

	for _, direction := range []string{"up", "down"} {
		template := "crud/migrations/__migration__." + direction + ".sql.tmpl"

		content, err := fs.ReadFile(templates.FS, template)
		if err != nil {
			return err
		}

		output, err := utils.Render(template, string(content), data)
		if err != nil {
			return err
		}

		if err := utils.WriteNewFile(project.Path("migrations", migration+"."+direction+".sql"), output); err != nil {
			return err
		}

		fmt.Println("created migrations/" + migration + "." + direction + ".sql")
	}

	return nil
}

func hasMigration(project *config.Project, names utils.Names) bool {

	files, _ := os.ReadDir(project.Path("migrations"))
	for _, file := range files {
		data, err := os.ReadFile(project.Path("migrations", file.Name()))
		if err == nil && strings.Contains(string(data), "CREATE TABLE IF NOT EXISTS "+names.Plural+" ") {
			return true
		}
	}

	return false
}

func crudDatabase(project *config.Project, folder string) string {

	data, err := os.ReadFile(project.Path("internal", "containers", folder, "repositories", "repository.go"))
	if err == nil && strings.Contains(string(data), `"database/sql"`) {
		return "postgres"
	}

	return "memory"
}

func containerCall(names utils.Names, db string) string {
	if db == "postgres" {
		return names.Plural + ".Init(ctx, api, pgdb)"
	}
	return names.Plural + ".Init(ctx, api)"
}
