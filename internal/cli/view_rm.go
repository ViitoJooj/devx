package cli

import (
	"fmt"
	"os"

	"github.com/ViitoJooj/devx/internal/cli/utils"
	"github.com/ViitoJooj/devx/pkg/errorx"
	"github.com/spf13/cobra"
)

var viewRmYes bool

var viewRmCmd = &cobra.Command{
	Use:     "rm <name>",
	Aliases: []string{"remove"},
	Short:   "Remove a frontend from www/",
	Args:    cobra.ExactArgs(1),
	Run:     viewRm,
}

func init() {
	viewRmCmd.Flags().BoolVarP(&viewRmYes, "yes", "y", false, "skip the confirmation")
}

func viewRm(cmd *cobra.Command, args []string) {

	project := loadAPIProject()

	var target *viewInfo
	installed := installedViews(project)
	for i := range installed {
		if installed[i].Name == args[0] {
			target = &installed[i]
		}
	}

	if target == nil {
		errorx.ViewNotFound(args[0])
	}

	confirm("Remove www/"+target.Name+" from "+project.Name+"?", viewRmYes)

	if err := os.RemoveAll(project.Path("www", target.Name)); err != nil {
		errorx.CannotGenerate(err)
	}
	utils.RemoveEmptyDirs(project.Root, project.Path("www"))

	var stillUsed []string
	for _, other := range installedViews(project) {
		stillUsed = append(stillUsed, views[other.Kind].origins(other.Port)...)
	}

	var unused []string
	for _, origin := range views[target.Kind].origins(target.Port) {
		if !contains(stillUsed, origin) {
			unused = append(unused, origin)
		}
	}

	if err := utils.RemoveEnvList(project.Path(".env"), "APPLICATION_FRONTEND", unused...); err != nil {
		errorx.CannotGenerate(err)
	}

	if err := syncMakefile(project); err != nil {
		errorx.CannotGenerate(err)
	}

	if len(installedViews(project)) == 0 {
		if err := removeCors(project); err != nil {
			errorx.CannotGenerate(err)
		}

		if err := utils.RunGo(project.Root, "mod", "tidy"); err != nil {
			errorx.CannotRunGo(err)
		}
	}

	fmt.Println("removed www/" + target.Name)
}
