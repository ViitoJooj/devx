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
	Short:     "Add the api, cli, mcp or jobs to an existing project",
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

	if project.Has(a.Name) {
		errorx.AppAlreadyExists(a.Name)
	}

	defer utils.Prefetch(appFetches(a.Name)...)()

	if project.Standalone() {
		if err := utils.Step("moving to the monorepo layout", func() error { return promote(project) }); err != nil {
			errorx.CannotGenerate(err)
		}
	}

	if err := installApp(project.Root, projectData(project), a); err != nil {
		errorx.CannotGenerate(err)
	}

	if err := utils.Step("go mod tidy", func() error { return utils.RunGo(project.Root, "mod", "tidy") }); err != nil {
		errorx.CannotRunGo(err)
	}

	fmt.Println(a.Name + " added in cmd/" + a.Name)
}
