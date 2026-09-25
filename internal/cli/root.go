package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "devx",
	Short: "dev helper to generate and evolve projects",
}

func Execute() {
	rootCmd.AddCommand(initCmd, appCmd, serviceCmd, crudCmd, workerCmd, viewCmd, commandCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Oops. An error while executing devx '%s'\n", err)
		os.Exit(1)
	}
}
