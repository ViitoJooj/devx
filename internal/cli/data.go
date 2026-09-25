package cli

import (
	"github.com/ViitoJooj/devx/internal/cli/config"
	"github.com/ViitoJooj/devx/internal/cli/utils"
	"github.com/ViitoJooj/devx/pkg/errorx"
)

type templateData struct {
	Module    string
	Name      string
	Migration string
	Port      string
	Interval  string
	GoVersion string
	DB        string
	utils.Names
}

func loadProject() *config.Project {
	project, err := config.LoadProject()
	if err != nil {
		errorx.NotInProject()
	}
	return project
}

// loadAPIProject is used by the commands that change cmd/api (service, crud, worker, view).
func loadAPIProject() *config.Project {
	project := loadProject()
	if !project.HasAPI() {
		errorx.MissingApp("api")
	}
	return project
}

// loadCLIProject is used by the commands that change cmd/cli (command).
func loadCLIProject() *config.Project {
	project := loadProject()
	if !project.HasCLI() {
		errorx.MissingApp("cli")
	}
	return project
}
