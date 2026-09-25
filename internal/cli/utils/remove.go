package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// RemoveSnippet deletes the lines injected by Inject, comparing lines without indentation and repeated spaces.
func RemoveSnippet(path, snippet string) error {

	if strings.TrimSpace(snippet) == "" {
		return nil
	}

	lines, err := readLines(path)
	if err != nil {
		return err
	}

	var target []string
	for line := range strings.SplitSeq(snippet, "\n") {
		if normalize(line) != "" {
			target = append(target, normalize(line))
		}
	}
	if len(target) == 0 {
		return nil
	}

	for i := range lines {
		end, ok := matchFrom(lines, i, target)
		if !ok {
			continue
		}

		lines = append(lines[:i], lines[end:]...)
		lines = dropDoubleBlank(lines, i)
		return writeFormatted(path, strings.Join(lines, "\n"))
	}

	return nil
}

// squeezeBlankLines drops leading blank lines and keeps at most one blank line in a row.
func squeezeBlankLines(lines []string) string {

	var output []string
	for _, line := range lines {
		blank := strings.TrimSpace(line) == ""
		if blank && (len(output) == 0 || strings.TrimSpace(output[len(output)-1]) == "") {
			continue
		}
		output = append(output, line)
	}

	return strings.TrimRight(strings.Join(output, "\n"), "\n") + "\n"
}

// RemoveEnvBlock deletes the header comment and the keys of an env snippet, keeping the rest of the file.
func RemoveEnvBlock(path, snippet string) error {

	lines, err := readLines(path)
	if err != nil {
		return err
	}

	remove := map[string]bool{}
	var header string
	for line := range strings.SplitSeq(snippet, "\n") {
		line = strings.TrimSpace(line)
		if key, _, ok := strings.Cut(line, "="); ok {
			remove[key] = true
		} else if strings.HasPrefix(line, "#") && header == "" {
			header = line
		}
	}

	var output []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		key, _, _ := strings.Cut(trimmed, "=")
		if (header != "" && trimmed == header) || remove[key] {
			continue
		}
		if trimmed == "" && len(output) > 0 && strings.TrimSpace(output[len(output)-1]) == "" {
			continue
		}
		output = append(output, line)
	}

	content := strings.TrimRight(strings.Join(output, "\n"), "\n") + "\n"
	return os.WriteFile(path, []byte(content), 0o644)
}

// RemoveEmptyDirs removes dir and its parents while they are empty, stopping at root.
func RemoveEmptyDirs(root, dir string) {
	for dir != root && strings.HasPrefix(dir, root) {
		if err := os.Remove(dir); err != nil {
			return
		}
		dir = filepath.Dir(dir)
	}
}

func matchFrom(lines []string, start int, target []string) (int, bool) {
	j := start
	for k := 0; k < len(target); {
		if j >= len(lines) {
			return 0, false
		}

		current := normalize(lines[j])
		switch {
		case current == "" && k > 0:
			j++
		case current == target[k]:
			j++
			k++
		default:
			return 0, false
		}
	}
	return j, true
}

func dropDoubleBlank(lines []string, i int) []string {
	if i < len(lines) && strings.TrimSpace(lines[i]) == "" && (i == 0 || strings.TrimSpace(lines[i-1]) == "") {
		return append(lines[:i], lines[i+1:]...)
	}
	return lines
}

func normalize(line string) string {
	return strings.Join(strings.Fields(line), " ")
}

func indentOf(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

func readLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(data), "\n"), nil
}
