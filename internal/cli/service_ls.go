package cli

import (
	"strings"

	"github.com/spf13/cobra"
)

var serviceLsJSON bool

var serviceLsCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List the services and which ones are in the project",
	Args:    cobra.NoArgs,
	Run:     serviceLs,
}

func init() {
	serviceLsCmd.Flags().BoolVar(&serviceLsJSON, "json", false, "output as json")
}

type serviceItem struct {
	Name     string   `json:"name"`
	App      string   `json:"app"`
	Added    bool     `json:"added"`
	Requires []string `json:"requires"`
}

func serviceLs(cmd *cobra.Command, args []string) {

	project := loadProject()

	items := make([]serviceItem, 0, len(serviceOrder))
	var rows [][]string

	for _, name := range serviceOrder {
		svc := services[name]

		item := serviceItem{
			Name:     name,
			App:      svc.app(),
			Added:    project.Exists(svc.detectPath()...),
			Requires: append([]string{}, svc.Requires...),
		}
		items = append(items, item)

		status := "available"
		switch {
		case item.Added:
			status = "added"
		case !project.Exists(svc.mainPath()...):
			status = "needs " + item.App
		}

		rows = append(rows, []string{name, item.App, status, strings.Join(item.Requires, ", ")})
	}

	printList(serviceLsJSON, items, []string{"NAME", "APP", "STATUS", "REQUIRES"}, rows)
}
