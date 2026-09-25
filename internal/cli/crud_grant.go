package cli

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ViitoJooj/devx/internal/cli/config"
	"github.com/ViitoJooj/devx/internal/cli/utils"
	"github.com/ViitoJooj/devx/pkg/errorx"
	"github.com/spf13/cobra"
)

var crudGrantCmd = &cobra.Command{
	Use:       "grant <name> <select|insert|update|delete...>",
	Short:     "Allow the api user to run operations in the table of an entity (creates a migration)",
	Args:      cobra.MinimumNArgs(2),
	ValidArgs: operations,
	Run: func(cmd *cobra.Command, args []string) {
		changeAccess(true, args)
	},
}

var crudRevokeCmd = &cobra.Command{
	Use:       "revoke <name> <select|insert|update|delete...>",
	Short:     "Block operations of the api user in the table of an entity (creates a migration)",
	Args:      cobra.MinimumNArgs(2),
	ValidArgs: operations,
	Run: func(cmd *cobra.Command, args []string) {
		changeAccess(false, args)
	},
}

func changeAccess(grant bool, args []string) {

	project := loadAPIProject()

	names := resolveCrud(project, args[0])

	if crudDatabase(project, names.FolderPlural) != "postgres" {
		errorx.CrudNotPostgres(names.FolderPlural)
	}

	access, _ := tableAccess(project)
	current, ok := access[names.Plural]
	if !ok {
		errorx.TableNotFound(names.Plural)
	}

	var changes []string
	for _, op := range args[1:] {
		op = strings.ToLower(op)
		if !slices.Contains(operations, op) {
			errorx.UnknownOperation(op)
		}

		if slices.Contains(current, op) != grant && !slices.Contains(changes, op) {
			changes = append(changes, op)
		}
	}

	if len(changes) == 0 {
		fmt.Println(names.Plural + " already has that access")
		return
	}

	if err := writeAccessMigration(project, names, grant, changes); err != nil {
		errorx.CannotGenerate(err)
	}
}

func writeAccessMigration(project *config.Project, names utils.Names, grant bool, changes []string) error {

	number, err := nextMigrationNumber(project)
	if err != nil {
		return err
	}

	action, undo := "grant", "revoke"
	if !grant {
		action, undo = "revoke", "grant"
	}

	name := fmt.Sprintf("%04d_%s_%s_%s", number, names.Pascal, action, strings.Join(changes, "_"))
	user := `"` + appUser(project) + `"`
	ops := strings.ToUpper(strings.Join(changes, ", "))

	statements := map[string]string{
		"grant":  "GRANT " + ops + " ON " + names.Plural + " TO " + user + ";\n",
		"revoke": "REVOKE " + ops + " ON " + names.Plural + " FROM " + user + ";\n",
	}

	files := map[string]string{
		name + ".up.sql":   statements[action],
		name + ".down.sql": statements[undo],
	}

	for file, content := range files {
		if err := utils.WriteNewFile(project.Path("migrations", file), content); err != nil {
			return err
		}
		fmt.Println("created migrations/" + file)
	}

	return nil
}
