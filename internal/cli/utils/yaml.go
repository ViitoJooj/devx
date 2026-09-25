package utils

import (
	"os"
	"strings"
)

type yamlBlock struct {
	start  int
	end    int
	indent int
}

// findYaml locates the block of a key path (ex: services, api, environment) by indentation.
func findYaml(lines []string, keys []string) (yamlBlock, bool) {

	block := yamlBlock{start: -1, end: len(lines), indent: -1}

	for _, key := range keys {
		found := false

		for i := block.start + 1; i < block.end; i++ {
			trimmed := strings.TrimSpace(lines[i])
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}

			indent := indentOf(lines[i])
			if indent <= block.indent {
				break
			}

			if indent == childIndent(lines, block) && (trimmed == key+":" || strings.HasPrefix(trimmed, key+": ")) {
				block = yamlBlock{start: i, end: blockEnd(lines, i, indent), indent: indent}
				found = true
				break
			}
		}

		if !found {
			return yamlBlock{}, false
		}
	}

	return block, true
}

func childIndent(lines []string, block yamlBlock) int {
	for i := block.start + 1; i < block.end; i++ {
		if strings.TrimSpace(lines[i]) != "" && !strings.HasPrefix(strings.TrimSpace(lines[i]), "#") {
			return indentOf(lines[i])
		}
	}
	return block.indent + 2
}

func blockEnd(lines []string, start, indent int) int {

	end := start + 1
	for end < len(lines) {
		trimmed := strings.TrimSpace(lines[end])
		if trimmed != "" && indentOf(lines[end]) <= indent {
			break
		}
		end++
	}

	for end > start+1 && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}

	return end
}

// YamlAppend adds snippet as the last child of the key path, creating missing keys; gap separates it with a blank line.
func YamlAppend(path string, keys []string, snippet string, gap bool) error {

	lines, err := readLines(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	lines = trimTrailingBlank(lines)

	if alreadyInjected(strings.Join(lines, "\n"), snippet) {
		return nil
	}

	for len(keys) > 0 {
		if _, ok := findYaml(lines, keys); ok {
			break
		}
		snippet = keys[len(keys)-1] + ":\n" + indentBlock(snippet, "  ")
		keys = keys[:len(keys)-1]
		gap = len(keys) == 0
	}

	insertAt := len(lines)
	indent := ""
	hasChildren := len(lines) > 0

	if len(keys) > 0 {
		block, _ := findYaml(lines, keys)
		insertAt = block.end
		indent = strings.Repeat(" ", childIndent(lines, block))
		hasChildren = block.end > block.start+1
	}

	block := strings.Split(indentBlock(strings.TrimRight(snippet, "\n"), indent), "\n")
	if gap && hasChildren {
		block = append([]string{""}, block...)
	}

	lines = append(lines[:insertAt:insertAt], append(block, lines[insertAt:]...)...)
	return os.WriteFile(path, []byte(squeezeBlankLines(lines)), 0o644)
}

// YamlRemove deletes the key path and everything nested under it.
func YamlRemove(path string, keys []string) error {

	lines, err := readLines(path)
	if err != nil {
		return err
	}

	block, ok := findYaml(lines, keys)
	if !ok {
		return nil
	}

	lines = append(lines[:block.start], lines[block.end:]...)
	return os.WriteFile(path, []byte(squeezeBlankLines(lines)), 0o644)
}

// YamlHasChildren reports whether the key path exists and has nested lines.
func YamlHasChildren(path string, keys []string) bool {

	lines, err := readLines(path)
	if err != nil {
		return false
	}

	block, ok := findYaml(lines, keys)
	return ok && block.end > block.start+1
}

func indentBlock(text, indent string) string {
	var output []string
	for line := range strings.SplitSeq(text, "\n") {
		if strings.TrimSpace(line) == "" {
			output = append(output, "")
			continue
		}
		output = append(output, indent+line)
	}
	return strings.Join(output, "\n")
}

func trimTrailingBlank(lines []string) []string {
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}
