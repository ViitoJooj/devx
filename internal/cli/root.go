package cli

import (
	"fmt"
	"os"

	"github.com/ViitoJooj/devx/internal/cli/utils"
	"github.com/ViitoJooj/devx/pkg/errorx"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "devx",
	Short: "dev helper to generate and evolve projects",
}

func Execute() {
	errorx.OnExit(utils.UnlockProject)

	rootCmd.AddCommand(initCmd, appCmd, serviceCmd, crudCmd, workerCmd, viewCmd, commandCmd, migrateCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Oops. An error while executing devx '%s'\n", err)
		errorx.Exit(1)
	}

	utils.UnlockProject()
}
