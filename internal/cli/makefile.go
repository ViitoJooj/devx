package cli

import (
	"slices"
	"strings"

	"github.com/ViitoJooj/devx/internal/cli/config"
	"github.com/ViitoJooj/devx/internal/cli/utils"
)

const composeCmd = "docker compose -f infra/compose/docker-compose.yaml --env-file .env"

var managedTargets = []string{"run-dev", "run-api", "run-build", "docker-restart", "run-docker", "sudo-clean", "run-cli", "build-cli"}

type makeRule struct {
	Targets string
	Content string
}

// syncMakefile rewrites the targets devx owns from the project state and keeps the ones written by hand.
func syncMakefile(project *config.Project) error {

	path := project.Path("makefile")
	rules := desiredRules(project)

	desired := map[string]bool{}
	for _, rule := range rules {
		desired[rule.Targets] = true
	}

	err := utils.MakeRemoveRules(path, func(targets string) bool {
		return managedTarget(targets) && !desired[targets]
	})
	if err != nil {
		return err
	}

	after := ""
	for _, rule := range rules {
		if err := utils.MakeSetRule(path, rule.Targets, after, rule.Content); err != nil {
			return err
		}
		after = rule.Targets
	}

	return nil
}

func managedTarget(targets string) bool {

	if slices.Contains(managedTargets, targets) || strings.HasPrefix(targets, "view-") {
		return true
	}

	for target := range strings.FieldsSeq(targets) {
		if !slices.Contains(cleanNames(), target) {
			return false
		}
	}
	return true
}

func cleanNames() []string {
	var names []string
	for _, name := range serviceOrder {
		if clean := services[name].Clean; clean != nil {
			names = append(names, clean.Name)
		}
	}
	return names
}

func desiredRules(project *config.Project) []makeRule {

	var rules []makeRule

	if project.HasAPI() {
		rules = append(rules, apiRules(project)...)
	}

	if project.HasCLI() {
		rules = append(rules,
			makeRule{"run-cli", "run-cli:\n\tgo run ./cmd/cli $(ARGS)"},
			makeRule{"build-cli", "build-cli:\n\tgo build -o ./tmp/" + project.Name + " ./cmd/cli"},
		)
	}

	return rules
}

func apiRules(project *config.Project) []makeRule {

	compose := project.Exists("infra", "compose", "docker-compose.yaml")
	installed := installedViews(project)

	runDev := "run-dev:"
	if compose {
		runDev += "\n\t" + composeCmd + " up -d"
	}

	var viewTargets []string
	for _, info := range installed {
		viewTargets = append(viewTargets, "view-"+info.Name)
	}

	if len(viewTargets) > 0 {
		runDev += "\n\t$(MAKE) -j run-api " + strings.Join(viewTargets, " ")
	} else {
		runDev += "\n\tgo tool air"
	}

	rules := []makeRule{{"run-dev", runDev}}

	if len(viewTargets) > 0 {
		rules = append(rules, makeRule{"run-api", "run-api:\n\tgo tool air"})
	}

	rules = append(rules, makeRule{"run-build", "run-build:\n\tgo build -o ./tmp/main ./cmd/api"})

	if compose {
		rules = append(rules, makeRule{"docker-restart", "docker-restart:\n\t" + composeCmd + " down\n\t" + composeCmd + " up -d"})
	}

	if project.Exists(dockerService.detectPath()...) {
		rules = append(rules, makeRule{"run-docker", "run-docker:\n\t" + composeCmd + " --profile app up -d --build"})
	}

	for _, info := range installed {
		rules = append(rules, makeRule{"view-" + info.Name, "view-" + info.Name + ":\n\tcd www/" + info.Name + " && bun install && " + views[info.Kind].Dev})
	}

	var targets []cleanTarget
	for _, name := range serviceOrder {
		svc := services[name]
		if svc.Clean != nil && project.Exists(svc.detectPath()...) {
			targets = append(targets, *svc.Clean)
		}
	}

	if len(targets) > 0 {
		var names []string
		for _, target := range targets {
			names = append(names, target.Name)
		}

		rules = append(rules,
			makeRule{"sudo-clean", sudoCleanRule(targets)},
			makeRule{strings.Join(names, " "), strings.Join(names, " ") + ":\n\t@:"},
		)
	}

	return rules
}

func sudoCleanRule(targets []cleanTarget) string {

	var names []string
	var cases strings.Builder

	for _, target := range targets {
		names = append(names, target.Name)

		cases.WriteString("\t\t\t" + target.Name + ") \\\n")
		for line := range strings.SplitSeq(target.Script, "\n") {
			cases.WriteString("\t\t\t\t" + line + "\n")
		}
		cases.WriteString("\t\t\t\t;; \\\n")
	}

	list := strings.Join(names, " ")
	options := strings.Join(names, "|")

	return "sudo-clean:\n" +
		"\t@targets=\"$(filter-out sudo-clean,$(MAKECMDGOALS))\"; \\\n" +
		"\tif [ -z \"$$targets\" ]; then \\\n" +
		"\t\techo \"uso: make sudo-clean " + options + " [" + list + "]\"; \\\n" +
		"\t\texit 1; \\\n" +
		"\tfi; \\\n" +
		"\tfor target in $$targets; do \\\n" +
		"\t\tcase \"$$target\" in \\\n" +
		cases.String() +
		"\t\t\t*) \\\n" +
		"\t\t\t\techo \"arg desconhecido: $$target (use " + options + ")\"; \\\n" +
		"\t\t\t\texit 1; \\\n" +
		"\t\t\t\t;; \\\n" +
		"\t\tesac; \\\n" +
		"\tdone"
}
