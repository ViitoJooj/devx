package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/ViitoJooj/devx/internal/cli/config"
	"github.com/ViitoJooj/devx/internal/cli/utils"
)

var operations = []string{"select", "insert", "update", "delete"}

var defaultOperations = []string{"select", "insert", "update"}

var (
	createTable = regexp.MustCompile(`(?i)CREATE TABLE IF NOT EXISTS (\w+)`)
	accessSQL   = regexp.MustCompile(`(?i)^(GRANT|REVOKE) ([A-Z, ]+) ON (\w+) (?:TO|FROM) `)
)

type migrationFile struct {
	name    string
	content string
}

func upMigrations(project *config.Project) []migrationFile {

	files, _ := filepath.Glob(project.Path("migrations", "*.up.sql"))
	sort.Strings(files)

	var output []migrationFile
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err == nil {
			output = append(output, migrationFile{name: filepath.Base(file), content: string(data)})
		}
	}

	return output
}

// tableAccess replays the GRANT/REVOKE migrations over the default privileges of init.sh.
func tableAccess(project *config.Project) (map[string][]string, []string) {

	access := map[string][]string{}
	var tables []string

	for _, migration := range upMigrations(project) {
		for _, match := range createTable.FindAllStringSubmatch(migration.content, -1) {
			table := strings.ToLower(match[1])
			if _, ok := access[table]; !ok {
				tables = append(tables, table)
				access[table] = slices.Clone(defaultOperations)
			}
		}

		for line := range strings.SplitSeq(migration.content, "\n") {
			match := accessSQL.FindStringSubmatch(strings.TrimSpace(line))
			if match == nil {
				continue
			}

			table := strings.ToLower(match[3])
			for op := range strings.SplitSeq(match[2], ",") {
				op = strings.ToLower(strings.TrimSpace(op))
				if strings.EqualFold(match[1], "GRANT") {
					if !slices.Contains(access[table], op) {
						access[table] = append(access[table], op)
					}
				} else {
					access[table] = slices.DeleteFunc(access[table], func(item string) bool { return item == op })
				}
			}
		}
	}

	return access, tables
}

func appUser(project *config.Project) string {

	data, err := os.ReadFile(project.Path(".env"))
	if err == nil {
		for line := range strings.SplitSeq(string(data), "\n") {
			if user, ok := strings.CutPrefix(strings.TrimSpace(line), "POSTGRES_APP_USER="); ok && user != "" {
				return user
			}
		}
	}

	return project.Name + "_app"
}

// crudAccess is the access column of crud ls: "-" for memory, "unknown" when no migration creates the table.
func crudAccess(project *config.Project, access map[string][]string, folder string, names utils.Names) ([]string, string) {

	if crudDatabase(project, folder) != "postgres" {
		return nil, "-"
	}

	ops, ok := access[names.Plural]
	if !ok {
		return nil, "unknown"
	}

	var ordered []string
	for _, op := range operations {
		if slices.Contains(ops, op) {
			ordered = append(ordered, op)
		}
	}

	if len(ordered) == 0 {
		return []string{}, "none"
	}

	return ordered, strings.Join(ordered, ",")
}
