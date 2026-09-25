package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ViitoJooj/devx/internal/cli/config"
	"github.com/ViitoJooj/devx/internal/cli/utils"
	"github.com/ViitoJooj/devx/pkg/errorx"
	"github.com/spf13/cobra"
)

var workerInterval string

var workerAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Create a worker in internal/workers and start it in main",
	Args:  cobra.ExactArgs(1),
	Run:   workerAdd,
}

func init() {
	workerAddCmd.Flags().StringVar(&workerInterval, "interval", "1h", "default interval between runs (30s, 15m, 1h...)")
}

func workerAdd(cmd *cobra.Command, args []string) {

	project := loadAPIProject()

	names, err := utils.NewNames(args[0])
	if err != nil {
		errorx.InvalidName(args[0])
	}

	interval, err := time.ParseDuration(workerInterval)
	if err != nil || interval <= 0 {
		errorx.InvalidInterval(workerInterval)
	}

	if project.Exists("internal", "workers", names.Folder) {
		errorx.WorkerAlreadyExists(names.Folder)
	}

	if names.Singular == "workers" || importedInMain(project, names.Singular) {
		errorx.PackageConflict(names.Singular)
	}

	created, err := utils.RenderDir(utils.RenderInput{
		Source:      "worker",
		Destination: project.Root,
		Data: templateData{
			Module:   project.Module,
			Name:     project.Name,
			Interval: utils.GoDuration(interval),
			Names:    names,
		},
		Placeholders: map[string]string{
			"__package__": names.Folder,
		},
		SkipExisting: true,
	})
	if err != nil {
		errorx.CannotGenerate(err)
	}

	if err := addWorkerToMain(project, names); err != nil {
		errorx.CannotGenerate(err)
	}

	for _, file := range created {
		fmt.Println("created " + file)
	}
}

func addWorkerToMain(project *config.Project, names utils.Names) error {

	path := project.Path("cmd", "api", "main.go")

	if err := addImport(path, workerImport(project, names)); err != nil {
		return err
	}

	if mainContains(project, "workers.New(") {
		return utils.InsertGo(path, callArgAnchor("workers.New"), workerStart(names))
	}

	if err := addImport(path, workersPackageImport(project)); err != nil {
		return err
	}

	return utils.InsertGo(path, workersAnchor, "workers.New(\n\t"+workerStart(names)+"\n).Start(ctx)\n\n")
}

func removeWorkersBlock(project *config.Project) error {

	if len(workersIn(project)) > 0 {
		return nil
	}

	path := project.Path("cmd", "api", "main.go")

	if err := utils.RemoveGoStmt(path, "main", "workers.New("); err != nil {
		return err
	}

	if err := utils.RemoveSnippet(path, workersPackageImport(project)); err != nil {
		return err
	}

	for _, file := range [][]string{{"internal", "workers", "workers.go"}, {"internal", "contracts", "worker.go"}} {
		if err := os.Remove(project.Path(file...)); err != nil && !os.IsNotExist(err) {
			return err
		}
		utils.RemoveEmptyDirs(project.Root, project.Path(file[:len(file)-1]...))
	}

	return nil
}

func workersPackageImport(project *config.Project) string {
	return `"` + project.Module + `/internal/workers"`
}

func mainContains(project *config.Project, text string) bool {
	main, err := os.ReadFile(project.Path("cmd", "api", "main.go"))
	if err != nil {
		return false
	}
	return strings.Contains(string(main), text)
}

func workerImport(project *config.Project, names utils.Names) string {
	return names.Singular + ` "` + project.Module + "/internal/workers/" + names.Folder + `"`
}

func workerStart(names utils.Names) string {
	return names.Singular + ".New" + names.Pascal + "Worker(),"
}

func importedInMain(project *config.Project, pkg string) bool {
	return mainContains(project, "/"+pkg+`"`) || mainContains(project, " "+pkg+" \"")
}
