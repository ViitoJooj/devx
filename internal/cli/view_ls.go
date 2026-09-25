package cli

import (
	"strconv"

	"github.com/spf13/cobra"
)

var viewLsJSON bool

var viewLsCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List the frontends in www/",
	Args:    cobra.NoArgs,
	Run:     viewLs,
}

func init() {
	viewLsCmd.Flags().BoolVar(&viewLsJSON, "json", false, "output as json")
}

func viewLs(cmd *cobra.Command, args []string) {

	project := loadAPIProject()

	items := installedViews(project)
	if items == nil {
		items = []viewInfo{}
	}

	var rows [][]string
	for _, item := range items {
		port := "-"
		if item.Port > 0 {
			port = strconv.Itoa(item.Port)
		}
		rows = append(rows, []string{item.Name, item.Kind, port, "www/" + item.Name})
	}

	printList(viewLsJSON, items, []string{"NAME", "VIEW", "PORT", "PATH"}, rows)
}
