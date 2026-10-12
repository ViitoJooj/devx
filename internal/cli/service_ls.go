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

// serviceInfo is what 'service ls' tells about each service.
var serviceInfo = map[string]struct{ Category, Description string }{
	"postgres":   {"database", "postgresql connection, migrations and the app user with minimum privileges"},
	"redis":      {"cache", "redis connection (cache, sessions, rate limit)"},
	"rabbitmq":   {"queue", "rabbitmq connection for async messages"},
	"resend":     {"email", "resend client to send emails"},
	"stripe":     {"payments", "stripe client for payments"},
	"prometheus": {"observability", "metrics middleware and the /metrics route"},
	"grafana":    {"observability", "grafana with the prometheus datasource (adds prometheus)"},
	"docker":     {"deploy", "Dockerfile and the api in the compose (profile app)"},
	"viper":      {"config", "config.yaml loaded with viper in the cli"},
}

type serviceItem struct {
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	App         string   `json:"app"`
	Added       bool     `json:"added"`
	Requires    []string `json:"requires"`
}

func serviceLs(cmd *cobra.Command, args []string) {

	project := loadProject()

	items := make([]serviceItem, 0, len(serviceOrder))
	var rows [][]string

	for _, name := range serviceOrder {
		svc := services[name]

		item := serviceItem{
			Name:        name,
			Category:    serviceInfo[name].Category,
			Description: serviceInfo[name].Description,
			App:         svc.app(),
			Added:       project.Exists(svc.detectPath()...),
			Requires:    append([]string{}, svc.Requires...),
		}
		items = append(items, item)

		status := "available"
		switch {
		case item.Added:
			status = "added"
		case !project.Exists(svc.mainPath()...):
			status = "needs " + item.App
		}

		rows = append(rows, []string{name, item.Category, item.App, status, strings.Join(item.Requires, ", "), item.Description})
	}

	printList(serviceLsJSON, items, []string{"NAME", "CATEGORY", "APP", "STATUS", "REQUIRES", "DESCRIPTION"}, rows)
}
