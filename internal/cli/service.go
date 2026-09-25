package cli

import (
	"github.com/spf13/cobra"
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage the project services (postgres, redis, docker...)",
}

func init() {
	serviceCmd.AddCommand(serviceAddCmd, serviceRmCmd, serviceLsCmd)
}
