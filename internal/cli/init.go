package cli

import (
	"fmt"
	"os"

	"github.com/ViitoJooj/devx/internal/cli/utils"
	"github.com/ViitoJooj/devx/pkg/errorx"
	"github.com/spf13/cobra"
)

const defaultModulePrefix = "github.com/ViitoJooj/"

var (
	initModule string
	initAPI    bool
	initCLI    bool
	initMCP    bool
	initJobs   bool
)

var initCmd = &cobra.Command{
	Use:   "init <project> [--api] [--cli] [--mcp] [--jobs]",
	Short: "Create a new project: an api (default), or a lean cli, mcp or jobs app; --api --cli together make a monorepo",
	Args:  cobra.ExactArgs(1),
	Run:   initProject,
}

func init() {
	initCmd.Flags().StringVar(&initModule, "module", "", "go module path (default "+defaultModulePrefix+"<project>)")
	initCmd.Flags().BoolVar(&initAPI, "api", false, "create the api (gin, noxacloud structure); default when no flag is passed")
	initCmd.Flags().BoolVar(&initCLI, "cli", false, "create the cli (cobra); alone it uses the lean layout (main.go + cmd/ + internal/)")
	initCmd.Flags().BoolVar(&initMCP, "mcp", false, "create an mcp server (lean layout: cmd/server.go, internal/tools, internal/resources)")
	initCmd.Flags().BoolVar(&initJobs, "jobs", false, "create a jobs runner (lean layout: cmd/worker.go, internal/jobs, internal/schedules)")
}

func initProject(cmd *cobra.Command, args []string) {

	if !initAPI && !initCLI && !initMCP && !initJobs {
		initAPI = true
	}

	selected := map[string]bool{"api": initAPI, "cli": initCLI, "mcp": initMCP, "jobs": initJobs}

	count := 0
	for _, on := range selected {
		if on {
			count++
		}
	}

	standalone := count == 1 && !initAPI

	name := args[0]
	if !utils.ValidName(name) {
		errorx.InvalidName(name)
	}

	if _, err := os.Stat(name); err == nil {
		errorx.ProjectAlreadyExists(name)
	}

	module := initModule
	if module == "" {
		module = defaultModulePrefix + name
	}

	var kinds []string
	for _, kind := range []string{"api", "cli", "mcp", "jobs"} {
		if selected[kind] {
			kinds = append(kinds, kind)
		}
	}

	// downloads run while the files are generated
	defer utils.Prefetch(appFetches(kinds...)...)()

	data := templateData{
		Module:     module,
		Name:       name,
		Port:       "8080",
		Standalone: standalone,
		API:        initAPI,
		CLI:        initCLI,
		MCP:        initMCP,
		Jobs:       initJobs,
	}

	err := utils.Step("creating base files", func() error {
		_, err := utils.RenderDir(utils.RenderInput{
			Source:      "base",
			Destination: name,
			Data:        data,
		})
		return err
	})
	if err != nil {
		os.RemoveAll(name)
		errorx.CannotGenerate(err)
	}

	if err := utils.Step("go mod init", func() error { return utils.RunGo(name, "mod", "init", module) }); err != nil {
		os.RemoveAll(name)
		errorx.CannotRunGo(err)
	}

	if standalone {
		kind := "cli"
		switch {
		case initMCP:
			kind = "mcp"
		case initJobs:
			kind = "jobs"
		}

		if err := utils.Step("installing "+kind, func() error { return installStandalone(name, data, kind) }); err != nil {
			os.RemoveAll(name)
			errorx.CannotGenerate(err)
		}
	}

	for _, kind := range appOrder {
		if standalone || !selected[kind] {
			continue
		}

		if err := utils.Step("installing "+kind, func() error { return installApp(name, data, apps[kind]) }); err != nil {
			os.RemoveAll(name)
			errorx.CannotGenerate(err)
		}
	}

	if err := utils.Step("go mod tidy", func() error { return utils.RunGo(name, "mod", "tidy") }); err != nil {
		os.RemoveAll(name)
		errorx.CannotRunGo(err)
	}

	fmt.Println("project " + name + " created")
	if initAPI {
		fmt.Println("api: cd " + name + " && devx service add postgres && devx crud add <name>")
	}
	if initCLI {
		fmt.Println("cli: cd " + name + " && devx command add <name> && make run-cli ARGS=\"<name>\"")
	}
	if initMCP {
		fmt.Println("mcp: cd " + name + " && make run-mcp")
	}
	if initJobs {
		fmt.Println("jobs: cd " + name + " && make run-jobs")
	}
}
