package utils

import (
	"os"
	"strings"
)

type makeRule struct {
	Targets string
	start   int
	end     int
}

func makeRules(lines []string) []makeRule {

	var rules []makeRule
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if line == "" || strings.HasPrefix(line, "\t") || strings.HasPrefix(line, "#") || strings.Contains(line, "=") {
			continue
		}

		targets, _, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}

		end := i + 1
		for end < len(lines) && strings.HasPrefix(lines[end], "\t") {
			end++
		}

		rules = append(rules, makeRule{Targets: strings.TrimSpace(targets), start: i, end: end})
		i = end - 1
	}

	return rules
}

// MakeTargets lists the targets line of every rule in the makefile.
func MakeTargets(path string) []string {

	lines, err := readLines(path)
	if err != nil {
		return nil
	}

	var output []string
	for _, rule := range makeRules(lines) {
		output = append(output, rule.Targets)
	}
	return output
}

// MakeSetRule replaces the rule with the same targets, or inserts it after the rule `after` (end of file if missing).
func MakeSetRule(path, targets, after, rule string) error {

	lines, err := readLines(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	ruleLines := strings.Split(strings.TrimRight(rule, "\n"), "\n")
	rules := makeRules(lines)

	for _, existing := range rules {
		if existing.Targets == targets {
			lines = append(lines[:existing.start:existing.start], append(ruleLines, lines[existing.end:]...)...)
			return os.WriteFile(path, []byte(squeezeBlankLines(lines)), 0o644)
		}
	}

	insertAt := -1
	for _, existing := range rules {
		if existing.Targets == after {
			insertAt = existing.end
		}
	}

	if insertAt < 0 {
		lines = append(trimTrailingBlank(lines), "")
		lines = append(lines, ruleLines...)
		return os.WriteFile(path, []byte(squeezeBlankLines(lines)), 0o644)
	}

	block := append([]string{""}, ruleLines...)
	lines = append(lines[:insertAt:insertAt], append(block, lines[insertAt:]...)...)
	return os.WriteFile(path, []byte(squeezeBlankLines(lines)), 0o644)
}

// MakeRemoveRules deletes every rule whose targets match.
func MakeRemoveRules(path string, match func(targets string) bool) error {

	lines, err := readLines(path)
	if err != nil {
		return err
	}

	rules := makeRules(lines)
	for i := len(rules) - 1; i >= 0; i-- {
		if match(rules[i].Targets) {
			lines = append(lines[:rules[i].start], lines[rules[i].end:]...)
		}
	}

	return os.WriteFile(path, []byte(squeezeBlankLines(lines)), 0o644)
}
