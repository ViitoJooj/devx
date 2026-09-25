package cli

import (
	"github.com/spf13/cobra"
)

var commandLsJSON bool

var commandLsCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List the cli commands in internal/commands",
	Args:    cobra.NoArgs,
	Run:     commandLs,
}

func init() {
	commandLsCmd.Flags().BoolVar(&commandLsJSON, "json", false, "output as json")
}

func commandLs(cmd *cobra.Command, args []string) {

	project := loadCLIProject()

	items := commandsIn(project)
	if items == nil {
		items = []commandItem{}
	}

	var rows [][]string
	for _, item := range items {
		rows = append(rows, []string{item.Name, item.Short, item.File})
	}

	printList(commandLsJSON, items, []string{"NAME", "SHORT", "FILE"}, rows)
}
