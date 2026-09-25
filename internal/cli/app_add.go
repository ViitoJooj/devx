package cli

import (
	"fmt"
	"strings"

	"github.com/ViitoJooj/devx/internal/cli/utils"
	"github.com/ViitoJooj/devx/pkg/errorx"
	"github.com/spf13/cobra"
)

var appAddCmd = &cobra.Command{
	Use:       "add <" + strings.Join(appOrder, "|") + ">",
	Short:     "Add the api or the cli to an existing project",
	Args:      cobra.ExactArgs(1),
	ValidArgs: appOrder,
	Run:       appAdd,
}

func appAdd(cmd *cobra.Command, args []string) {

	project := loadProject()

	a, ok := apps[args[0]]
	if !ok {
		errorx.InvalidName(args[0])
	}

	if project.Exists("cmd", a.Name, "main.go") {
		errorx.AppAlreadyExists(a.Name)
	}

	if err := installApp(project.Root, projectData(project), a); err != nil {
		errorx.CannotGenerate(err)
	}

	if err := utils.RunGo(project.Root, "mod", "tidy"); err != nil {
		errorx.CannotRunGo(err)
	}

	fmt.Println(a.Name + " added in cmd/" + a.Name)
}
