package cli

import (
	"fmt"
	"os"

	"github.com/ViitoJooj/devx/internal/cli/utils"
	"github.com/ViitoJooj/devx/pkg/errorx"
	"github.com/spf13/cobra"
)

var workerRmYes bool

var workerRmCmd = &cobra.Command{
	Use:     "rm <name>",
	Aliases: []string{"remove"},
	Short:   "Remove a worker and stop starting it in main",
	Args:    cobra.ExactArgs(1),
	Run:     workerRm,
}

func init() {
	workerRmCmd.Flags().BoolVarP(&workerRmYes, "yes", "y", false, "skip the confirmation")
}

func workerRm(cmd *cobra.Command, args []string) {

	project := loadAPIProject()

	names, err := utils.NewNames(args[0])
	if err != nil {
		errorx.InvalidName(args[0])
	}

	if !project.Exists("internal", "workers", names.Folder) {
		errorx.WorkerNotFound(names.Folder)
	}

	confirm("Remove worker "+names.Folder+" from "+project.Name+"?", workerRmYes)

	if err := os.RemoveAll(project.Path("internal", "workers", names.Folder)); err != nil {
		errorx.CannotGenerate(err)
	}

	main := project.Path("cmd", "api", "main.go")

	if err := utils.RemoveSnippet(main, workerStart(names)); err != nil {
		errorx.CannotGenerate(err)
	}

	if err := utils.RemoveSnippet(main, workerImport(project, names)); err != nil {
		errorx.CannotGenerate(err)
	}

	if err := removeWorkersBlock(project); err != nil {
		errorx.CannotGenerate(err)
	}

	fmt.Println("removed internal/workers/" + names.Folder)
}
