package cli

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ViitoJooj/devx/internal/cli/config"
	"github.com/ViitoJooj/devx/internal/cli/utils"
	"github.com/spf13/cobra"
)

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "Manage the project apps (cmd/api, cmd/cli, cmd/mcp, cmd/jobs)",
}

func init() {
	appCmd.AddCommand(appAddCmd, appRmCmd, appLsCmd)
}

type app struct {
	Name string
	Deps []string
}

var apps = map[string]app{
	"api": {
		Name: "api",
		Deps: []string{"github.com/gin-gonic/gin", "github.com/joho/godotenv"},
	},
	"cli": {
		Name: "cli",
		Deps: []string{"github.com/spf13/cobra"},
	},
	"mcp": {
		Name: "mcp",
		Deps: []string{"github.com/modelcontextprotocol/go-sdk"},
	},
	"jobs": {
		Name: "jobs",
		Deps: []string{"github.com/robfig/cron/v3"},
	},
}

var appOrder = []string{"api", "cli", "mcp", "jobs"}

// appFetches lists what installing the apps downloads, for Prefetch.
func appFetches(kinds ...string) []utils.Fetch {

	var fetches []utils.Fetch
	for _, kind := range kinds {
		for _, dep := range apps[kind].Deps {
			fetches = append(fetches, utils.Fetch{Path: dep})
		}
	}
	return fetches
}

func installApp(root string, data templateData, a app) error {

	var err error
	switch a.Name {
	case "mcp", "jobs":
		err = installLean(root, data, a.Name)
	default:
		_, err = utils.RenderDir(utils.RenderInput{
			Source:       a.Name,
			Destination:  root,
			Data:         data,
			SkipExisting: true,
		})
	}
	if err != nil {
		return err
	}

	project := &config.Project{Root: root, Module: data.Module, Name: data.Name}
	if err := syncMakefile(project); err != nil {
		return err
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

// installLean adds mcp or jobs to a monorepo: the lean template is rendered aside and given the monorepo layout,
// so there is a single template per app.
func installLean(root string, data templateData, kind string) error {

	tmp, err := os.MkdirTemp("", "devx-app-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	if _, err := utils.RenderDir(utils.RenderInput{Source: "standalone/" + kind, Destination: tmp, Data: data}); err != nil {
		return err
	}

	aside := &config.Project{Root: tmp, Module: data.Module, Name: data.Name}

	if kind == "mcp" {
		err = promoteMCP(aside)
	} else {
		err = promoteJobs(aside)
	}
	if err != nil {
		return err
	}

	return filepath.WalkDir(tmp, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		rel, err := filepath.Rel(tmp, path)
		if err != nil {
			return err
		}

		dest := filepath.Join(root, rel)
		if _, err := os.Stat(dest); err == nil {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return utils.WriteNewFile(dest, string(content))
	})
}
