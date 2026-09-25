package utils

import (
	"fmt"
	"go/format"
	"os"
	"strings"
)

func AppendFile(path, content string) error {

	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if alreadyInjected(string(data), content) {
		return nil
	}

	output := string(data)
	if output != "" && !strings.HasSuffix(output, "\n\n") {
		output = strings.TrimRight(output, "\n") + "\n\n"
	}

	return os.WriteFile(path, []byte(output+content), 0o644)
}

func alreadyInjected(content, snippet string) bool {

	var target []string
	for line := range strings.SplitSeq(snippet, "\n") {
		if normalize(line) != "" {
			target = append(target, normalize(line))
		}
	}
	if len(target) == 0 {
		return true
	}

	lines := strings.Split(content, "\n")
	for i := range lines {
		if _, ok := matchFrom(lines, i, target); ok {
			return true
		}
	}
	return false
}

func writeFormatted(path, content string) error {

	if strings.HasSuffix(path, ".go") {
		formatted, err := format.Source([]byte(content))
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		content = string(formatted)
	}

	return os.WriteFile(path, []byte(content), 0o644)
}
