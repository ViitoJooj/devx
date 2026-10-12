package cli

import (
	"go/format"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ViitoJooj/devx/internal/cli/config"
	"github.com/ViitoJooj/devx/internal/cli/utils"
)

// promote turns the lean layout (main.go in the root) into the monorepo one (cmd/<app>), so other apps can live beside it.
func promote(project *config.Project) error {

	switch {
	case project.HasCLI():
		return promoteCLI(project)
	case project.HasMCP():
		return promoteMCP(project)
	case project.HasJobs():
		return promoteJobs(project)
	}

	return nil
}

// registerCommand puts the command in the register call of the root file.
func registerCommand(project *config.Project, names utils.Names) error {

	root := rootFile(project)

	if rootContains(project, registerCall(project)+"(") {
		return utils.InsertGo(root, callArgAnchor(registerCall(project)), commandRegister(names))
	}

	return utils.InsertGo(root, registerAnchor(project), registerCall(project)+"(\n\t"+commandRegister(names)+"\n)\n")
}

// promoteCLI moves the root to internal/commands/root.go and keeps every command registered.
func promoteCLI(project *config.Project) error {

	var registered []utils.Names
	for _, item := range commandsIn(project) {
		names, err := utils.NewNames(strings.TrimSuffix(filepath.Base(item.File), ".go"))
		if err != nil {
			continue
		}
		registered = append(registered, names)
	}

	for _, path := range []string{"main.go", "cmd/root.go", "internal/commands/register.go"} {
		if err := os.Remove(project.Path(filepath.FromSlash(path))); err != nil {
			return err
		}
	}

	if _, err := utils.RenderDir(utils.RenderInput{
		Source:      "cli",
		Destination: project.Root,
		Data:        projectData(project),
	}); err != nil {
		return err
	}

	// the lean layout is gone, so the monorepo root is used from here
	for _, names := range registered {
		if err := registerCommand(project, names); err != nil {
			return err
		}
	}

	utils.RemoveEmptyDirs(project.Root, project.Path("cmd"))
	return nil
}

// promoteMCP moves the server to internal/agentics, with its tools and resources.
func promoteMCP(project *config.Project) error {

	module := project.Module

	if err := moveFile(project, "cmd/server.go", "internal/agentics/server.go", func(src string) string {
		src = strings.Replace(src, "package cmd", "package agentics", 1)
		return strings.Replace(src, "func Execute()", "func Run()", 1)
	}); err != nil {
		return err
	}

	for _, dir := range []string{"tools", "resources"} {
		if err := os.Rename(project.Path("internal", dir), project.Path("internal", "agentics", dir)); err != nil {
			return err
		}

		if err := rewriteImports(project.Root, module+"/internal/"+dir+`"`, module+"/internal/agentics/"+dir+`"`); err != nil {
			return err
		}
	}

	return writeMain(project, "mcp", "agentics")
}

// promoteJobs moves the runner to internal/schedules/run.go, beside the schedules.
func promoteJobs(project *config.Project) error {

	module := project.Module

	if err := moveFile(project, "cmd/worker.go", "internal/schedules/run.go", func(src string) string {
		src = strings.Replace(src, "package cmd", "package schedules", 1)
		src = strings.Replace(src, "func Execute()", "func Run()", 1)
		src = strings.Replace(src, "\t\""+module+"/internal/schedules\"\n", "", 1)
		return strings.ReplaceAll(src, "schedules.Register(", "Register(")
	}); err != nil {
		return err
	}

	return writeMain(project, "jobs", "schedules")
}

// writeMain replaces the root main.go with cmd/<app>/main.go calling pkg.Run().
func writeMain(project *config.Project, app, pkg string) error {

	if err := os.Remove(project.Path("main.go")); err != nil {
		return err
	}

	main := "package main\n\nimport \"" + project.Module + "/internal/" + pkg + "\"\n\nfunc main() {\n\t" + pkg + ".Run()\n}\n"

	if err := utils.WriteNewFile(project.Path("cmd", app, "main.go"), main); err != nil {
		return err
	}

	utils.RemoveEmptyDirs(project.Root, project.Path("cmd"))
	return nil
}

func moveFile(project *config.Project, from, to string, change func(string) string) error {

	data, err := os.ReadFile(project.Path(filepath.FromSlash(from)))
	if err != nil {
		return err
	}

	formatted, err := format.Source([]byte(change(string(data))))
	if err != nil {
		return err
	}

	if err := utils.WriteNewFile(project.Path(filepath.FromSlash(to)), string(formatted)); err != nil {
		return err
	}

	return os.Remove(project.Path(filepath.FromSlash(from)))
}

// rewriteImports changes an import path in every go file of the project.
func rewriteImports(root, old, replacement string) error {
	return rewriteGo(root, func(_, src string) string {
		return strings.ReplaceAll(src, old, replacement)
	})
}

// rewriteGo applies change to every go file of the project and keeps the ones it changed formatted (imports sorted).
func rewriteGo(root string, change func(path, src string) string) error {

	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && (d.Name() == ".git" || d.Name() == "node_modules" || d.Name() == "tmp") {
			return filepath.SkipDir
		}

		if d.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		changed := change(path, string(data))
		if changed == string(data) {
			return nil
		}

		if formatted, err := format.Source([]byte(changed)); err == nil {
			changed = string(formatted)
		}

		return os.WriteFile(path, []byte(changed), 0o644)
	})
}
