package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/ViitoJooj/devx/internal/cli/config"
)

var tauriOrigins = []string{"tauri://localhost", "http://tauri.localhost"}

var devScript = regexp.MustCompile(`("dev":\s*")([^"]*)(")`)

type view struct {
	Kind string
	// Scaffold returns the bunx args to create the view in www/<name>.
	Scaffold func(name, identifier string) []string
	// BasePort is the first port tried; FixedPort views (tauri + vite) can not change it.
	BasePort  int
	FixedPort bool
	PortFlag  string
	Dev       string
	Tauri     bool
}

var views = map[string]view{
	"tauri": {
		Kind: "tauri",
		Scaffold: func(name, identifier string) []string {
			return []string{"create-tauri-app@latest", name, "--template", "vanilla", "--manager", "bun", "--identifier", identifier, "--yes"}
		},
		Dev:   "bun run tauri dev",
		Tauri: true,
	},
	"tauri-svelte": {
		Kind: "tauri-svelte",
		Scaffold: func(name, identifier string) []string {
			return []string{"create-tauri-app@latest", name, "--template", "svelte-ts", "--manager", "bun", "--identifier", identifier, "--yes"}
		},
		BasePort:  1420,
		FixedPort: true,
		Dev:       "bun run tauri dev",
		Tauri:     true,
	},
	"tauri-react": {
		Kind: "tauri-react",
		Scaffold: func(name, identifier string) []string {
			return []string{"create-tauri-app@latest", name, "--template", "react-ts", "--manager", "bun", "--identifier", identifier, "--yes"}
		},
		BasePort:  1420,
		FixedPort: true,
		Dev:       "bun run tauri dev",
		Tauri:     true,
	},
	"sveltekit": {
		Kind: "sveltekit",
		Scaffold: func(name, _ string) []string {
			return []string{"sv@latest", "create", name, "--template", "minimal", "--types", "ts", "--no-add-ons", "--install", "bun"}
		},
		BasePort: 5173,
		PortFlag: " --port %d --strictPort",
		Dev:      "bun run dev",
	},
	"react": {
		Kind: "react",
		Scaffold: func(name, _ string) []string {
			return []string{"create-vite@latest", name, "--template", "react-ts", "--no-interactive", "--no-immediate"}
		},
		BasePort: 5173,
		PortFlag: " --port %d --strictPort",
		Dev:      "bun run dev",
	},
	"next": {
		Kind: "next",
		Scaffold: func(name, _ string) []string {
			return []string{"create-next-app@latest", name, "--ts", "--app", "--src-dir", "--eslint", "--tailwind", "--use-bun", "--disable-git", "--yes"}
		},
		BasePort: 3001,
		PortFlag: " -p %d",
		Dev:      "bun run dev",
	},
}

var viewOrder = []string{"tauri", "tauri-svelte", "tauri-react", "sveltekit", "react", "next"}

// origins are the browser origins that call the API from this view.
func (v view) origins(port int) []string {
	var output []string
	if port > 0 {
		output = append(output, "http://localhost:"+strconv.Itoa(port))
	}
	if v.Tauri {
		output = append(output, tauriOrigins...)
	}
	return output
}

type viewInfo struct {
	Name string `json:"name"`
	Kind string `json:"view"`
	Port int    `json:"port"`
}

type packageJSON struct {
	Devx viewInfo `json:"devx"`
}

// installedViews reads the devx field written by view add in www/*/package.json.
func installedViews(project *config.Project) []viewInfo {

	entries, _ := os.ReadDir(project.Path("www"))

	var output []viewInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		data, err := os.ReadFile(project.Path("www", entry.Name(), "package.json"))
		if err != nil {
			continue
		}

		var pkg packageJSON
		if err := json.Unmarshal(data, &pkg); err != nil || pkg.Devx.Kind == "" {
			continue
		}

		pkg.Devx.Name = entry.Name()
		output = append(output, pkg.Devx)
	}

	return output
}

func nextViewPort(project *config.Project, v view) int {

	if v.FixedPort || v.BasePort == 0 {
		return v.BasePort
	}

	used := map[int]bool{}
	for _, installed := range installedViews(project) {
		used[installed.Port] = true
	}

	port := v.BasePort
	for used[port] {
		port++
	}

	return port
}

// tagPackage stores the view info in package.json and pins the dev port.
func tagPackage(path string, v view, port int) error {

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)

	if v.PortFlag != "" {
		content = devScript.ReplaceAllString(content, "${1}${2}"+fmt.Sprintf(v.PortFlag, port)+"${3}")
	}

	tag := fmt.Sprintf(`"devx": { "view": %q, "port": %d },`, v.Kind, port)
	content = strings.Replace(content, "{", "{\n\t"+tag, 1)

	return os.WriteFile(path, []byte(content), 0o644)
}
