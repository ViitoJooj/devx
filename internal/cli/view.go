package cli

import (
	"os"
	"strings"

	"github.com/ViitoJooj/devx/internal/cli/config"
	"github.com/ViitoJooj/devx/internal/cli/utils"
	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Manage the project frontends in www/ (" + strings.Join(viewOrder, ", ") + ")",
}

func init() {
	viewCmd.AddCommand(viewAddCmd, viewRmCmd, viewLsCmd)
}

const corsSnippet = `api.Use(httpx.Cors(cfg.Application.FrontendURL))
`

func addCors(project *config.Project) error {

	if project.Exists("pkg", "httpx", "cors.go") {
		return nil
	}

	_, err := utils.RenderDir(utils.RenderInput{
		Source:       "views",
		Destination:  project.Path("pkg", "httpx"),
		Data:         projectData(project),
		SkipExisting: true,
	})
	if err != nil {
		return err
	}

	if err := utils.InsertGo(project.Path("cmd", "api", "main.go"), connAnchor, corsSnippet+"\n"); err != nil {
		return err
	}

	return utils.RunGo(project.Root, "get", "github.com/gin-contrib/cors")
}

func removeCors(project *config.Project) error {

	if err := utils.RemoveSnippet(project.Path("cmd", "api", "main.go"), corsSnippet); err != nil {
		return err
	}

	if err := os.Remove(project.Path("pkg", "httpx", "cors.go")); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func excludeWwwFromAir(project *config.Project) error {

	path := project.Path(".air.toml")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	content := string(data)
	if strings.Contains(content, `"www"`) {
		return nil
	}

	content = strings.Replace(content, "exclude_dir = [", `exclude_dir = ["www", `, 1)
	return os.WriteFile(path, []byte(content), 0o644)
}
