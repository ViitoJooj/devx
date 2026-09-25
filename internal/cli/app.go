package cli

import (
	"github.com/ViitoJooj/devx/internal/cli/config"
	"github.com/ViitoJooj/devx/internal/cli/utils"
	"github.com/spf13/cobra"
)

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "Manage the project apps (api in cmd/api, cli in cmd/cli)",
}

func init() {
	appCmd.AddCommand(appAddCmd, appRmCmd, appLsCmd)
}

type app struct {
	Name  string
	Tools []string
	Deps  []string
}

var apps = map[string]app{
	"api": {
		Name:  "api",
		Tools: []string{"github.com/air-verse/air@latest"},
	},
	"cli": {
		Name: "cli",
		Deps: []string{"github.com/spf13/cobra"},
	},
}

var appOrder = []string{"api", "cli"}

func installApp(root string, data templateData, a app) error {

	_, err := utils.RenderDir(utils.RenderInput{
		Source:       a.Name,
		Destination:  root,
		Data:         data,
		SkipExisting: true,
	})
	if err != nil {
		return err
	}

	project := &config.Project{Root: root, Module: data.Module, Name: data.Name}
	if err := syncMakefile(project); err != nil {
		return err
	}

	for _, tool := range a.Tools {
		if err := utils.RunGo(root, "get", "-tool", tool); err != nil {
			return err
		}
	}

	for _, dep := range a.Deps {
		if err := utils.RunGo(root, "get", dep); err != nil {
			return err
		}
	}

	return nil
}
