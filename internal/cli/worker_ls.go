package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ViitoJooj/devx/internal/cli/config"
	"github.com/spf13/cobra"
)

var intervalExpr = regexp.MustCompile(`NewWorkerOptions\((\d+)\s*\*\s*time\.(\w+)`)

var workerUnits = map[string]time.Duration{
	"Millisecond": time.Millisecond,
	"Second":      time.Second,
	"Minute":      time.Minute,
	"Hour":        time.Hour,
}

var workerLsJSON bool

var workerLsCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List the workers in internal/workers",
	Args:    cobra.NoArgs,
	Run:     workerLs,
}

func init() {
	workerLsCmd.Flags().BoolVar(&workerLsJSON, "json", false, "output as json")
}

type workerItem struct {
	Name     string `json:"name"`
	Interval string `json:"interval"`
}

func workerLs(cmd *cobra.Command, args []string) {

	project := loadAPIProject()

	items := make([]workerItem, 0)
	var rows [][]string

	for _, name := range workersIn(project) {
		item := workerItem{Name: name, Interval: readInterval(project, name)}
		items = append(items, item)
		rows = append(rows, []string{item.Name, item.Interval})
	}

	printList(workerLsJSON, items, []string{"NAME", "INTERVAL"}, rows)
}

func workersIn(project *config.Project) []string {

	entries, _ := os.ReadDir(project.Path("internal", "workers"))

	var output []string
	for _, entry := range entries {
		if entry.IsDir() {
			output = append(output, entry.Name())
		}
	}

	return output
}

// readInterval reads the default interval written by worker add; "custom" when it was changed by hand.
func readInterval(project *config.Project, name string) string {

	files, _ := filepath.Glob(project.Path("internal", "workers", name, "*.go"))

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		match := intervalExpr.FindStringSubmatch(string(data))
		if match == nil {
			continue
		}

		amount, _ := strconv.Atoi(match[1])
		return shortDuration(time.Duration(amount) * workerUnits[match[2]])
	}

	return "custom"
}

func shortDuration(d time.Duration) string {
	output := d.String()
	if strings.HasSuffix(output, "m0s") {
		output = strings.TrimSuffix(output, "0s")
	}
	if strings.HasSuffix(output, "h0m") {
		output = strings.TrimSuffix(output, "0m")
	}
	return output
}
