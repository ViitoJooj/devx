package cli

import (
	"fmt"
	"os"

	"github.com/ViitoJooj/devx/internal/cli/utils"
	"github.com/ViitoJooj/devx/pkg/errorx"
	"github.com/spf13/cobra"
)

var commandRmYes bool

var commandRmCmd = &cobra.Command{
	Use:     "rm <name>",
	Aliases: []string{"remove"},
	Short:   "Remove a cli command and its register in root",
	Args:    cobra.ExactArgs(1),
	Run:     commandRm,
}

func init() {
	commandRmCmd.Flags().BoolVarP(&commandRmYes, "yes", "y", false, "skip the confirmation")
}

func commandRm(cmd *cobra.Command, args []string) {

	project := loadCLIProject()

	names, err := utils.NewNames(args[0])
	if err != nil || reservedCommand(names) || !project.Exists("internal", "commands", names.Singular+".go") {
		errorx.CommandNotFound(args[0])
	}

	confirm("Remove command "+names.Kebab+" from "+project.Name+"?", commandRmYes)

	if err := os.Remove(commandFile(project, names)); err != nil {
		errorx.CannotGenerate(err)
	}

	root := rootFile(project)

	if err := utils.RemoveSnippet(root, commandRegister(names)); err != nil {
		errorx.CannotGenerate(err)
	}

	if len(commandsIn(project)) == 0 {
		if err := utils.RemoveGoStmt(root, "Execute", "rootCmd.AddCommand("); err != nil {
			errorx.CannotGenerate(err)
		}
	}

	fmt.Println("removed internal/commands/" + names.Singular + ".go")
}
