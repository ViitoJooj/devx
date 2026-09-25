package cli

import (
	"github.com/spf13/cobra"
)

var appLsJSON bool

var appLsCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List the apps and which ones are in the project",
	Args:    cobra.NoArgs,
	Run:     appLs,
}

func init() {
	appLsCmd.Flags().BoolVar(&appLsJSON, "json", false, "output as json")
}

type appItem struct {
	Name  string `json:"name"`
	Added bool   `json:"added"`
	Path  string `json:"path"`
}

func appLs(cmd *cobra.Command, args []string) {

	project := loadProject()

	items := make([]appItem, 0, len(appOrder))
	var rows [][]string

	for _, name := range appOrder {
		item := appItem{Name: name, Added: project.Exists("cmd", name, "main.go"), Path: "cmd/" + name}
		items = append(items, item)

		status := "available"
		if item.Added {
			status = "added"
		}

		rows = append(rows, []string{item.Name, status, item.Path})
	}

	printList(appLsJSON, items, []string{"NAME", "STATUS", "PATH"}, rows)
}
