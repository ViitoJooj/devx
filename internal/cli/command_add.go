package cli

import (
	"fmt"

	"github.com/ViitoJooj/devx/internal/cli/utils"
	"github.com/ViitoJooj/devx/pkg/errorx"
	"github.com/spf13/cobra"
)

var commandAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Create a cobra command in internal/commands and register it in root",
	Args:  cobra.ExactArgs(1),
	Run:   commandAdd,
}

func commandAdd(cmd *cobra.Command, args []string) {

	project := loadCLIProject()

	names, err := utils.NewNames(args[0])
	if err != nil {
		errorx.InvalidName(args[0])
	}

	if reservedCommand(names) {
		errorx.ReservedName(names.Kebab)
	}

	if project.Exists("internal", "commands", names.Singular+".go") {
		errorx.CommandAlreadyExists(names.Kebab)
	}

	_, err = utils.RenderDir(utils.RenderInput{
		Source:       "command",
		Destination:  project.Root,
		Data:         templateData{Module: project.Module, Name: project.Name, Names: names},
		Placeholders: map[string]string{"__command__": names.Singular},
	})
	if err != nil {
		errorx.CannotGenerate(err)
	}

	root := rootFile(project)

	if rootContains(project, "rootCmd.AddCommand(") {
		err = utils.InsertGo(root, callArgAnchor("rootCmd.AddCommand"), commandRegister(names))
	} else {
		err = utils.InsertGo(root, beforeStmtAnchor("Execute", "if err := rootCmd.Execute()"), "rootCmd.AddCommand(\n\t"+commandRegister(names)+"\n)\n\n")
	}
	if err != nil {
		errorx.CannotGenerate(err)
	}

	fmt.Println("created internal/commands/" + names.Singular + ".go")
	fmt.Println("run with: make run-cli ARGS=\"" + names.Kebab + "\"")
}
