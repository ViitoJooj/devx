package cli

import (
	"github.com/spf13/cobra"
)

var crudCmd = &cobra.Command{
	Use:   "crud",
	Short: "Manage the project cruds (internal/containers)",
}

func init() {
	crudCmd.AddCommand(crudAddCmd, crudRmCmd, crudLsCmd, crudDBCmd, crudGrantCmd, crudRevokeCmd)
}
