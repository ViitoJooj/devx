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
		"cmd/api", "internal/restful", "internal/contracts", "internal/workers", "internal/config",
		"docs/openapi", "hacks/http", "infra", "www",
		".env", ".env.example", "Dockerfile", ".dockerignore",
		"internal/shared/httpx", "internal/shared/logs", "internal/shared/utils", "internal/shared/errorx/runtimeErrors.go",
		"pkg/database", "pkg/cache", "pkg/rabbitmq", "pkg/resend", "pkg/stripe",
	},
	"cli": {
		"cmd/cli", "internal/commands", "config.yaml", "internal/shared/errorx/config_errors.go",
	},
}

func appRm(cmd *cobra.Command, args []string) {

	project := loadProject()

	a, ok := apps[args[0]]
	if !ok {
		errorx.InvalidName(args[0])
	}

	if !project.Has(a.Name) {
		errorx.MissingApp(a.Name)
	}

	if !appRmSudo {
		errorx.SudoRequired("devx app rm " + a.Name)
	}

	if project.Standalone() || countApps(project) < 2 {
		errorx.LastApp(a.Name)
	}

	confirm("Remove the "+a.Name+" and everything in "+strings.Join(appPaths[a.Name], ", ")+"?", appRmYes)

	if err := removeApp(project, a); err != nil {
		errorx.CannotGenerate(err)
	}

	if err := utils.Step("go mod tidy", func() error { return utils.RunGo(project.Root, "mod", "tidy") }); err != nil {
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

func countApps(project *config.Project) int {

	count := 0
	for _, name := range []string{"api", "cli", "mcp", "jobs"} {
		if project.Has(name) {
			count++
		}
	}
	return count
}
