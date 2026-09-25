package cli

import (
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/ViitoJooj/devx/internal/cli/config"
	"github.com/ViitoJooj/devx/internal/cli/utils"
	"github.com/spf13/cobra"
)

var commandCmd = &cobra.Command{
	Use:   "command",
	Short: "Manage the cli commands (internal/commands)",
}

func init() {
	commandCmd.AddCommand(commandAddCmd, commandRmCmd, commandLsCmd)
}

var (
	commandUse   = regexp.MustCompile(`Use:\s*"([^"]+)"`)
	commandShort = regexp.MustCompile(`Short:\s*"([^"]*)"`)
)

var reservedCommands = []string{"root", "init", "main", "execute", "help", "completion", "version"}

func reservedCommand(names utils.Names) bool {
	return slices.Contains(reservedCommands, names.Singular) || token.IsKeyword(names.Camel)
}

func commandFile(project *config.Project, names utils.Names) string {
	return project.Path("internal", "commands", names.Singular+".go")
}

func commandRegister(names utils.Names) string {
	return names.Camel + "Cmd,"
}

type commandItem struct {
	Name  string `json:"name"`
	Short string `json:"short"`
	File  string `json:"file"`
}

// commandsIn reads the commands created by command add (every file in internal/commands but root.go).
func commandsIn(project *config.Project) []commandItem {

	files, _ := filepath.Glob(project.Path("internal", "commands", "*.go"))

	var output []commandItem
	for _, file := range files {
		if filepath.Base(file) == "root.go" {
			continue
		}

		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		use := commandUse.FindStringSubmatch(string(data))
		if use == nil {
			continue
		}

		item := commandItem{
			Name: strings.Fields(use[1])[0],
			File: "internal/commands/" + filepath.Base(file),
		}
		if short := commandShort.FindStringSubmatch(string(data)); short != nil {
			item.Short = short[1]
		}

		output = append(output, item)
	}

	return output
}

func rootFile(project *config.Project) string {
	return project.Path("internal", "commands", "root.go")
}

func rootContains(project *config.Project, text string) bool {
	data, err := os.ReadFile(rootFile(project))
	return err == nil && strings.Contains(string(data), text)
}
