package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/ViitoJooj/devx/internal/cli/utils"
	"github.com/ViitoJooj/devx/pkg/errorx"
	"github.com/spf13/cobra"
)

var nonAlnum = regexp.MustCompile(`[^a-z0-9]`)

var viewAddName string

var viewAddCmd = &cobra.Command{
	Use:       "add <" + strings.Join(viewOrder, "|") + ">",
	Short:     "Create a frontend in www/<name> with the official CLI (bun)",
	Args:      cobra.ExactArgs(1),
	ValidArgs: viewOrder,
	Run:       viewAdd,
}

func init() {
	viewAddCmd.Flags().StringVar(&viewAddName, "name", "", "folder in www/ (default: the view type)")
}

func viewAdd(cmd *cobra.Command, args []string) {

	project := loadAPIProject()

	v, ok := views[args[0]]
	if !ok {
		errorx.UnknownView(args[0])
	}

	name := viewAddName
	if name == "" {
		name = v.Kind
	}
	if !utils.ValidName(name) {
		errorx.InvalidName(name)
	}

	if project.Exists("www", name) {
		errorx.ViewAlreadyExists(name)
	}

	if _, err := exec.LookPath("bun"); err != nil {
		errorx.BunNotFound()
	}

	port := nextViewPort(project, v)
	for _, installed := range installedViews(project) {
		if v.FixedPort && installed.Port == port {
			fmt.Println("warning: www/" + installed.Name + " also uses port " + strconv.Itoa(port) + ", run only one of them at a time")
		}
	}

	if err := os.MkdirAll(project.Path("www"), 0o755); err != nil {
		errorx.CannotGenerate(err)
	}

	fmt.Println("creating www/" + name + " with " + v.Kind + "...")

	identifier := "com." + nonAlnum.ReplaceAllString(project.Name, "") + "." + nonAlnum.ReplaceAllString(name, "")

	if err := scaffoldView(project.Path("www"), name, v.Scaffold(name, identifier)); err != nil {
		os.RemoveAll(project.Path("www", name))
		utils.RemoveEmptyDirs(project.Root, project.Path("www"))
		errorx.CannotRunTool(err)
	}

	if err := tagPackage(project.Path("www", name, "package.json"), v, port); err != nil {
		errorx.CannotGenerate(err)
	}

	if err := utils.AddEnvList(project.Path(".env"), "APPLICATION_FRONTEND", v.origins(port)...); err != nil {
		errorx.CannotGenerate(err)
	}

	if err := addCors(project); err != nil {
		errorx.CannotGenerate(err)
	}

	if err := excludeWwwFromAir(project); err != nil {
		errorx.CannotGenerate(err)
	}

	if err := syncMakefile(project); err != nil {
		errorx.CannotGenerate(err)
	}

	if err := utils.RunGo(project.Root, "mod", "tidy"); err != nil {
		errorx.CannotRunGo(err)
	}

	fmt.Println("created www/" + name)
	if port > 0 {
		fmt.Println("dev server: http://localhost:" + strconv.Itoa(port))
	}
	fmt.Println("run with: make run-dev (api + views) or make view-" + name)
}

func scaffoldView(www, name string, args []string) error {

	if err := utils.Run(www, "bunx", args...); err != nil {
		return err
	}

	return utils.Run(filepath.Join(www, name), "bun", "install")
}
