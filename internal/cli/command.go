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

var reservedCommands = []string{"root", "init", "main", "execute", "help", "completion", "version", "register"}

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
		if base := filepath.Base(file); base == "root.go" || base == "register.go" {
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

// rootFile is where commands are registered: root.go in the monorepo, register.go in the lean layout.
func rootFile(project *config.Project) string {
	if project.Standalone() {
		return project.Path("internal", "commands", "register.go")
	}
	return project.Path("internal", "commands", "root.go")
}

// registerCall is the AddCommand call that holds the commands, on the root var of the file.
func registerCall(project *config.Project) string {
	if project.Standalone() {
		return "root.AddCommand"
	}
	return "rootCmd.AddCommand"
}

// registerAnchor is where the first AddCommand call goes: before the root.Execute in the monorepo, at the end of Register in the lean layout.
func registerAnchor(project *config.Project) utils.Anchor {
	if !project.Standalone() {
		return beforeStmtAnchor("Execute", "if err := rootCmd.Execute()")
	}

	return func(g *utils.GoFile) (int, string, error) {
		fn := g.Func("Register")
		if fn == nil || fn.Body == nil {
			return 0, "", anchorError(g, "func Register")
		}

		line := g.Line(fn.Body.Rbrace)
		return line, g.IndentOf(line) + "\t", nil
	}
}

// registerFunc is the function that holds the AddCommand call.
func registerFunc(project *config.Project) string {
	if project.Standalone() {
		return "Register"
	}
	return "Execute"
}

func rootContains(project *config.Project, text string) bool {
	data, err := os.ReadFile(rootFile(project))
	return err == nil && strings.Contains(string(data), text)
}
