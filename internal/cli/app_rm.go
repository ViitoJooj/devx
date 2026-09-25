package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ViitoJooj/devx/internal/cli/config"
	"github.com/ViitoJooj/devx/internal/cli/utils"
	"github.com/ViitoJooj/devx/pkg/errorx"
	"github.com/spf13/cobra"
)

var (
	appRmSudo bool
	appRmYes  bool
)

var appRmCmd = &cobra.Command{
	Use:       "rm <" + strings.Join(appOrder, "|") + "> --sudo",
	Aliases:   []string{"remove"},
	Short:     "Remove the api or the cli and everything that belongs to it (migrations are kept)",
	Args:      cobra.ExactArgs(1),
	ValidArgs: appOrder,
	Run:       appRm,
}

func init() {
	appRmCmd.Flags().BoolVar(&appRmSudo, "sudo", false, "required: confirms that the whole app will be deleted")
	appRmCmd.Flags().BoolVarP(&appRmYes, "yes", "y", false, "skip the confirmation")
}

var appPaths = map[string][]string{
	"api": {
		"cmd/api", "internal/containers", "internal/contracts", "internal/workers",
		"docs/openapi", "hacks", "infra", "www",
		".air.toml", ".env", ".env.example", "Dockerfile", ".dockerignore",
		"pkg/dotenv", "pkg/httpx", "pkg/logs", "pkg/utils", "pkg/errorx/runtimeErrors.go",
		"pkg/postgres", "pkg/migration", "pkg/redis", "pkg/rabbitmq", "pkg/resend", "pkg/stripe",
	},
	"cli": {
		"cmd/cli", "internal/commands", "config.yaml", "pkg/errorx/config_errors.go",
	},
}

func appRm(cmd *cobra.Command, args []string) {

	project := loadProject()

	a, ok := apps[args[0]]
	if !ok {
		errorx.InvalidName(args[0])
	}

	if !project.Exists("cmd", a.Name, "main.go") {
		errorx.MissingApp(a.Name)
	}

	if !appRmSudo {
		errorx.SudoRequired("devx app rm " + a.Name)
	}

	if !project.HasAPI() || !project.HasCLI() {
		errorx.LastApp(a.Name)
	}

	confirm("Remove the "+a.Name+" and everything in "+strings.Join(appPaths[a.Name], ", ")+"?", appRmYes)

	if err := removeApp(project, a); err != nil {
		errorx.CannotGenerate(err)
	}

	if err := utils.RunGo(project.Root, "mod", "tidy"); err != nil {
		errorx.CannotRunGo(err)
	}

	fmt.Println("removed the " + a.Name)
	if a.Name == "api" && project.Exists("migrations") {
		fmt.Println("migrations were kept")
	}
}

func removeApp(project *config.Project, a app) error {

	for _, path := range appPaths[a.Name] {
		full := project.Path(filepath.FromSlash(path))

		if err := os.RemoveAll(full); err != nil {
			return err
		}

		utils.RemoveEmptyDirs(project.Root, filepath.Dir(full))
	}

	return syncMakefile(project)
}
