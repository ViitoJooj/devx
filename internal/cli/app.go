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
		Deps:  []string{"github.com/gin-gonic/gin", "github.com/joho/godotenv"},
	},
	"cli": {
		Name: "cli",
		Deps: []string{"github.com/spf13/cobra"},
	},
}

// standaloneDeps are the modules the lean layouts import; go mod tidy resolves them, the prefetch only warms the cache.
var standaloneDeps = map[string][]string{
	"cli":  {"github.com/spf13/cobra"},
	"mcp":  {"github.com/modelcontextprotocol/go-sdk"},
	"jobs": {"github.com/robfig/cron/v3"},
}

// appFetches lists what installing the apps downloads, for Prefetch.
func appFetches(kinds ...string) []utils.Fetch {

	var fetches []utils.Fetch
	for _, kind := range kinds {
		for _, tool := range apps[kind].Tools {
			fetches = append(fetches, utils.Fetch{Path: tool, Tool: true})
		}
		for _, dep := range apps[kind].Deps {
			fetches = append(fetches, utils.Fetch{Path: dep})
		}
		for _, dep := range standaloneDeps[kind] {
			if _, monorepoApp := apps[kind]; !monorepoApp {
				fetches = append(fetches, utils.Fetch{Path: dep})
			}
		}
	}
	return fetches
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

	if len(a.Tools) > 0 {
		if err := utils.RunGo(root, append([]string{"get", "-tool"}, a.Tools...)...); err != nil {
			return err
		}
	}

	if len(a.Deps) > 0 {
		if err := utils.RunGo(root, append([]string{"get"}, a.Deps...)...); err != nil {
			return err
		}
	}

	return nil
}

// installStandalone renders the lean layout of a single app (main.go in the root) and its makefile targets.
func installStandalone(root string, data templateData, kind string) error {

	_, err := utils.RenderDir(utils.RenderInput{
		Source:      "standalone/" + kind,
		Destination: root,
		Data:        data,
	})
	if err != nil {
		return err
	}

	return syncMakefile(&config.Project{Root: root, Module: data.Module, Name: data.Name})
}
