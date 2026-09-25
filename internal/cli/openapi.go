package cli

import (
	"fmt"
	"io/fs"

	"github.com/ViitoJooj/devx/internal/cli/config"
	"github.com/ViitoJooj/devx/internal/cli/utils"
	"github.com/ViitoJooj/devx/internal/templates"
)

func openAPIPath(project *config.Project) string {
	return project.Path("docs", "openapi", "openapi.yaml")
}

func addOpenAPI(project *config.Project, names utils.Names) error {

	if !project.Exists("docs", "openapi", "openapi.yaml") {
		fmt.Println("docs/openapi/openapi.yaml not found, skipping openapi")
		return nil
	}

	injections := []struct {
		template string
		keys     []string
	}{
		{"snippets/openapi_paths.yaml.tmpl", []string{"paths"}},
		{"snippets/openapi_schemas.yaml.tmpl", []string{"components", "schemas"}},
	}

	for _, injection := range injections {
		content, err := fs.ReadFile(templates.FS, injection.template)
		if err != nil {
			return err
		}

		snippet, err := utils.Render(injection.template, string(content), templateData{Names: names})
		if err != nil {
			return err
		}

		if err := utils.YamlAppend(openAPIPath(project), injection.keys, snippet, false); err != nil {
			return err
		}
	}

	return nil
}

func removeOpenAPI(project *config.Project, names utils.Names) error {

	if !project.Exists("docs", "openapi", "openapi.yaml") {
		return nil
	}

	keys := [][]string{
		{"paths", "/v1/" + names.KebabPlural},
		{"paths", "/v1/" + names.KebabPlural + "/{id}"},
		{"components", "schemas", names.Pascal},
		{"components", "schemas", names.Pascal + "Input"},
		{"components", "schemas", names.Pascal + "Response"},
		{"components", "schemas", names.Pascal + "ListResponse"},
	}

	for _, key := range keys {
		if err := utils.YamlRemove(openAPIPath(project), key); err != nil {
			return err
		}
	}

	return nil
}

func documented(project *config.Project, kebabPlural string) bool {
	return utils.YamlHasChildren(openAPIPath(project), []string{"paths", "/v1/" + kebabPlural})
}
