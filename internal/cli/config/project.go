package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultPort      = "8080"
	defaultGoVersion = "1.27"
)

type Project struct {
	Root   string
	Module string
	Name   string
}

func LoadProject() (*Project, error) {

	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	goMod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return nil, err
	}

	module := moduleName(string(goMod))
	if module == "" {
		return nil, errors.New("module not found in go.mod")
	}

	project := &Project{
		Root:   root,
		Module: module,
		Name:   filepath.Base(module),
	}

	if !project.HasAPI() && !project.HasCLI() {
		return nil, errors.New("no cmd/api or cmd/cli")
	}

	return project, nil
}

func (p *Project) HasAPI() bool {
	return p.Exists("cmd", "api", "main.go")
}

func (p *Project) HasCLI() bool {
	return p.Exists("cmd", "cli", "main.go")
}

func (p *Project) Path(parts ...string) string {
	return filepath.Join(append([]string{p.Root}, parts...)...)
}

func (p *Project) Exists(parts ...string) bool {
	_, err := os.Stat(p.Path(parts...))
	return err == nil
}

func (p *Project) Port() string {

	env, err := os.ReadFile(p.Path(".env"))
	if err != nil {
		return defaultPort
	}

	for line := range strings.SplitSeq(string(env), "\n") {
		if port, ok := strings.CutPrefix(strings.TrimSpace(line), "APPLICATION_PORT="); ok && port != "" {
			return port
		}
	}

	return defaultPort
}

func (p *Project) GoVersion() string {

	goMod, err := os.ReadFile(p.Path("go.mod"))
	if err != nil {
		return defaultGoVersion
	}

	for line := range strings.SplitSeq(string(goMod), "\n") {
		if version, ok := strings.CutPrefix(strings.TrimSpace(line), "go "); ok {
			parts := strings.Split(version, ".")
			if len(parts) >= 2 {
				return parts[0] + "." + parts[1]
			}
			return version
		}
	}

	return defaultGoVersion
}

func moduleName(goMod string) string {
	for line := range strings.SplitSeq(goMod, "\n") {
		line = strings.TrimSpace(line)
		if module, ok := strings.CutPrefix(line, "module "); ok {
			return strings.TrimSpace(module)
		}
	}
	return ""
}
