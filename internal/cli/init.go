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
)

var initCmd = &cobra.Command{
	Use:   "init <project> [--api] [--cli]",
	Short: "Create a new project with an api (cmd/api, default), a cli (cmd/cli) or both",
	Args:  cobra.ExactArgs(1),
	Run:   initProject,
}

func init() {
	initCmd.Flags().StringVar(&initModule, "module", "", "go module path (default "+defaultModulePrefix+"<project>)")
	initCmd.Flags().BoolVar(&initAPI, "api", false, "create the api (gin, noxacloud structure); default when no flag is passed")
	initCmd.Flags().BoolVar(&initCLI, "cli", false, "create the cli (cobra + viper)")
}

func initProject(cmd *cobra.Command, args []string) {

	if !initAPI && !initCLI {
		initAPI = true
	}

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

	data := templateData{Module: module, Name: name, Port: "8080"}

	_, err := utils.RenderDir(utils.RenderInput{
		Source:      "base",
		Destination: name,
		Data:        data,
	})
	if err != nil {
		os.RemoveAll(name)
		errorx.CannotGenerate(err)
	}

	if err := utils.RunGo(name, "mod", "init", module); err != nil {
		os.RemoveAll(name)
		errorx.CannotRunGo(err)
	}

	for _, kind := range appOrder {
		if (kind == "api" && !initAPI) || (kind == "cli" && !initCLI) {
			continue
		}

		if err := installApp(name, data, apps[kind]); err != nil {
			os.RemoveAll(name)
			errorx.CannotGenerate(err)
		}
	}

	if err := utils.RunGo(name, "mod", "tidy"); err != nil {
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
}
