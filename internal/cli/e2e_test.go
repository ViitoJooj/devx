//go:build e2e

// End to end tests: they build devx, generate real projects and compile them, so they need the go tool and the module cache or network.
// Run with: go test -tags e2e -timeout 20m ./internal/cli/
package cli_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

var devx string

func TestMain(m *testing.M) {

	dir, err := os.MkdirTemp("", "devx-e2e-")
	if err != nil {
		panic(err)
	}

	devx = filepath.Join(dir, "devx")
	build := exec.Command("go", "build", "-o", devx, "../../cmd")
	if output, err := build.CombinedOutput(); err != nil {
		panic(string(output))
	}

	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func run(t *testing.T, dir string, name string, args ...string) string {
	t.Helper()

	cmd := exec.Command(name, args...)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s: %v\n%s", name, strings.Join(args, " "), err, output)
	}

	return string(output)
}

// fails runs a command that must be refused and returns its output.
func fails(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command(devx, args...)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("devx %s should fail\n%s", strings.Join(args, " "), output)
	}

	return string(output)
}

func newProject(t *testing.T, flags ...string) string {
	t.Helper()

	root := t.TempDir()
	run(t, root, devx, append([]string{"init", "demo"}, flags...)...)

	project := filepath.Join(root, "demo")
	run(t, project, "go", "vet", "./...")

	return project
}

func TestInitCombinations(t *testing.T) {
	for _, flags := range [][]string{
		{"--api"},
		{"--cli"},
		{"--mcp"},
		{"--jobs"},
		{"--api", "--cli"},
		{"--api", "--cli", "--mcp", "--jobs"},
	} {
		t.Run(strings.Join(flags, ""), func(t *testing.T) {
			newProject(t, flags...)
		})
	}
}

func TestCommandNames(t *testing.T) {
	project := newProject(t, "--cli")

	// names that used to be refused (init) or would clash with go (keywords, main)
	for _, name := range []string{"init", "type", "main", "execute", "go"} {
		run(t, project, devx, "command", "add", name)
	}

	run(t, project, "go", "vet", "./...")
	run(t, project, "go", "run", ".", "init")

	for _, name := range []string{"root", "register", "help"} {
		fails(t, project, "command", "add", name)
	}

	run(t, project, devx, "command", "rm", "init", "-y")
	run(t, project, "go", "vet", "./...")
}

func TestRunFromSubfolder(t *testing.T) {
	project := newProject(t, "--api")

	run(t, filepath.Join(project, "internal", "restful"), devx, "crud", "add", "user", "--db", "memory")
	run(t, project, "go", "vet", "./...")
}

func TestCrudReservedNames(t *testing.T) {
	project := newProject(t, "--api")

	for _, name := range []string{"type", "worker", "redis", "server", "config"} {
		fails(t, project, "crud", "add", name, "--db", "memory")
	}

	run(t, project, devx, "crud", "add", "user", "--db", "memory")
	run(t, project, "go", "vet", "./...")
}

func TestParallelRuns(t *testing.T) {
	project := newProject(t, "--api")

	var wg sync.WaitGroup
	for _, name := range []string{"alpha", "beta", "gamma", "delta"} {
		wg.Go(func() {
			exec.Command(devx, "crud", "add", name, "--db", "memory").Run()
		})
	}
	wg.Wait()

	run(t, project, "go", "vet", "./...")

	list := run(t, project, devx, "crud", "ls")
	for _, name := range []string{"alphas", "betas", "gammas", "deltas"} {
		for line := range strings.SplitSeq(list, "\n") {
			if strings.HasPrefix(line, name) && !strings.Contains(line, "yes") {
				t.Errorf("%s lost its openapi entry (lost update):\n%s", name, list)
			}
		}
	}
}

func TestPromotionKeepsCommands(t *testing.T) {
	project := newProject(t, "--cli")

	run(t, project, devx, "command", "add", "hello")
	run(t, project, devx, "app", "add", "api")
	run(t, project, devx, "app", "add", "mcp")
	run(t, project, devx, "app", "add", "jobs")
	run(t, project, "go", "vet", "./...")

	if out := run(t, project, "go", "run", "./cmd/cli", "hello"); !strings.Contains(out, "hello called") {
		t.Errorf("command lost in the promotion: %s", out)
	}
}

func TestMigrateOldLayout(t *testing.T) {
	project := newProject(t, "--api")

	// fake an old project: move the folders back and rewrite nothing; devx must refuse and point to migrate
	if err := os.MkdirAll(filepath.Join(project, "internal", "containers"), 0o755); err != nil {
		t.Fatal(err)
	}

	if out := fails(t, project, "crud", "ls"); !strings.Contains(out, "devx migrate") {
		t.Errorf("old layout should point to devx migrate:\n%s", out)
	}
}
