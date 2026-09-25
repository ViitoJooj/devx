package cli

import (
	"github.com/spf13/cobra"
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Manage the project workers (internal/workers)",
}

func init() {
	workerCmd.AddCommand(workerAddCmd, workerRmCmd, workerLsCmd)
}
