package utils

import (
	"os"
	"slices"
	"strings"
)

// AddEnvList appends values to a comma separated KEY=a,b in the env file, without duplicates.
func AddEnvList(path, key string, values ...string) error {
	return editEnvList(path, key, func(current []string) []string {
		for _, value := range values {
			if !slices.Contains(current, value) {
				current = append(current, value)
			}
		}
		return current
	})
}

// RemoveEnvList removes values from a comma separated KEY=a,b in the env file.
func RemoveEnvList(path, key string, values ...string) error {
	return editEnvList(path, key, func(current []string) []string {
		return slices.DeleteFunc(current, func(item string) bool {
			return slices.Contains(values, item)
		})
	})
}

func editEnvList(path, key string, edit func([]string) []string) error {

	lines, err := readLines(path)
	if err != nil {
		return err
	}

	for i, line := range lines {
		value, ok := strings.CutPrefix(strings.TrimSpace(line), key+"=")
		if !ok {
			continue
		}

		var current []string
		for item := range strings.SplitSeq(value, ",") {
			if item = strings.TrimSpace(item); item != "" {
				current = append(current, item)
			}
		}

		lines[i] = key + "=" + strings.Join(edit(current), ",")
		return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644)
	}

	lines = append(lines[:len(lines)-1], key+"="+strings.Join(edit(nil), ","), "")
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644)
}
