package config

import (
	"os"
	"path/filepath"
	"testing"
)

func touch(t *testing.T, root string, parts ...string) {
	t.Helper()
	path := filepath.Join(append([]string{root}, parts...)...)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestAppDetection(t *testing.T) {
	tests := []struct {
		name       string
		files      [][]string
		standalone bool
		has        map[string]bool
	}{
		{"monorepo", [][]string{{"cmd", "api", "main.go"}, {"cmd", "cli", "main.go"}}, false, map[string]bool{"api": true, "cli": true}},
		{"standalone cli", [][]string{{"main.go"}, {"cmd", "root.go"}}, true, map[string]bool{"cli": true}},
		{"standalone mcp", [][]string{{"main.go"}, {"cmd", "server.go"}}, true, map[string]bool{"mcp": true}},
		{"standalone jobs", [][]string{{"main.go"}, {"cmd", "worker.go"}}, true, map[string]bool{"jobs": true}},
		{"cmd file without main.go is not an app", [][]string{{"cmd", "root.go"}}, false, map[string]bool{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			for _, file := range tt.files {
				touch(t, root, file...)
			}

			p := &Project{Root: root}
			if p.Standalone() != tt.standalone {
				t.Errorf("Standalone() = %v, want %v", p.Standalone(), tt.standalone)
			}
			for _, app := range []string{"api", "cli", "mcp", "jobs"} {
				if p.Has(app) != tt.has[app] {
					t.Errorf("Has(%q) = %v, want %v", app, p.Has(app), tt.has[app])
				}
			}
		})
	}
}

func TestFindRoot(t *testing.T) {
	root := t.TempDir()
	touch(t, root, "go.mod")
	touch(t, root, "internal", "restful", "users", "x.go")

	got, err := FindRoot(filepath.Join(root, "internal", "restful", "users"))
	if err != nil || got != root {
		t.Fatalf("FindRoot() = %q, %v; want %q", got, err, root)
	}

	if _, err := FindRoot(t.TempDir()); err == nil {
		t.Fatal("FindRoot() without go.mod should fail")
	}
}
