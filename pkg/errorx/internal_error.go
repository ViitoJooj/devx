package errorx

import (
	"fmt"
	"os"
	"strings"
)

func NotInProject() {
	fmt.Println("code: dvx-00")
	fmt.Println("error: current folder is not a devx project")
	fmt.Println("error_type: user error.")
	fmt.Println("message: run this command in the project root or create one with 'devx init <project>'")
	os.Exit(1)
}

func ProjectAlreadyExists(name string) {
	fmt.Println("code: dvx-01")
	fmt.Println("error: folder '" + name + "' already exists")
	fmt.Println("error_type: user error.")
	fmt.Println("message: choose another project name or remove the folder")
	os.Exit(1)
}

func InvalidName(name string) {
	fmt.Println("code: dvx-02")
	fmt.Println("error: invalid name '" + name + "'")
	fmt.Println("error_type: user error.")
	fmt.Println("message: use only lowercase letters, numbers, '_' or '-', starting with a letter")
	os.Exit(1)
}

func UnknownService(name string) {
	fmt.Println("code: dvx-03")
	fmt.Println("error: unknown service '" + name + "'")
	fmt.Println("error_type: user error.")
	fmt.Println("message: try this options(postgres, redis, rabbitmq, resend, stripe, prometheus, grafana, docker, viper)")
	os.Exit(1)
}

func ContainerAlreadyExists(name string) {
	fmt.Println("code: dvx-04")
	fmt.Println("error: container '" + name + "' already exists")
	fmt.Println("error_type: user error.")
	fmt.Println("message: remove internal/containers/" + name + " or choose another name")
	os.Exit(1)
}

func CannotGenerate(err error) {
	fmt.Println("code: dvx-05")
	fmt.Println("error: error on generate files")
	fmt.Println("error_type: internal error.")
	fmt.Println("message: " + err.Error())
	if strings.Contains(err.Error(), "marker") || strings.Contains(err.Error(), "block") {
		fmt.Println("hint: projects created by older devx versions may miss devx markers (cmd/api/main.go, makefile...): compare with a fresh 'devx init'")
	}
	os.Exit(1)
}

func CannotRunGo(err error) {
	fmt.Println("code: dvx-06")
	fmt.Println("error: error on run go command")
	fmt.Println("error_type: internal error.")
	fmt.Println("message: verify your go installation and network. " + err.Error())
	os.Exit(1)
}

func WorkerAlreadyExists(name string) {
	fmt.Println("code: dvx-07")
	fmt.Println("error: worker '" + name + "' already exists")
	fmt.Println("error_type: user error.")
	fmt.Println("message: remove internal/workers/" + name + " or choose another name")
	os.Exit(1)
}

func InvalidInterval(interval string) {
	fmt.Println("code: dvx-08")
	fmt.Println("error: invalid interval '" + interval + "'")
	fmt.Println("error_type: user error.")
	fmt.Println("message: use a go duration like 30s, 15m, 1h or 24h")
	os.Exit(1)
}

func PackageConflict(name string) {
	fmt.Println("code: dvx-09")
	fmt.Println("error: package '" + name + "' is already imported in cmd/api/main.go")
	fmt.Println("error_type: user error.")
	fmt.Println("message: choose another name")
	os.Exit(1)
}

func ContainerNotFound(name string) {
	fmt.Println("code: dvx-10")
	fmt.Println("error: container '" + name + "' not found")
	fmt.Println("error_type: user error.")
	fmt.Println("message: list your cruds with 'devx crud ls'")
	os.Exit(1)
}

func ServiceNotAdded(name string) {
	fmt.Println("code: dvx-11")
	fmt.Println("error: service '" + name + "' is not in this project")
	fmt.Println("error_type: user error.")
	fmt.Println("message: see the project services with 'devx service ls'")
	os.Exit(1)
}

func ServiceInUse(name, usedBy string) {
	fmt.Println("code: dvx-12")
	fmt.Println("error: service '" + name + "' is used by " + usedBy)
	fmt.Println("error_type: user error.")
	fmt.Println("message: remove " + usedBy + " first")
	os.Exit(1)
}

func WorkerNotFound(name string) {
	fmt.Println("code: dvx-13")
	fmt.Println("error: worker '" + name + "' not found")
	fmt.Println("error_type: user error.")
	fmt.Println("message: list your workers with 'devx worker ls'")
	os.Exit(1)
}

func NeedConfirmation() {
	fmt.Println("code: dvx-14")
	fmt.Println("error: confirmation required")
	fmt.Println("error_type: user error.")
	fmt.Println("message: not running interactively, pass --yes to confirm")
	os.Exit(1)
}

func UnknownView(name string) {
	fmt.Println("code: dvx-15")
	fmt.Println("error: unknown view '" + name + "'")
	fmt.Println("error_type: user error.")
	fmt.Println("message: try this options(tauri, tauri-svelte, tauri-react, sveltekit, react, next)")
	os.Exit(1)
}

func ViewAlreadyExists(name string) {
	fmt.Println("code: dvx-16")
	fmt.Println("error: www/" + name + " already exists")
	fmt.Println("error_type: user error.")
	fmt.Println("message: use --name to create it in another folder")
	os.Exit(1)
}

func ViewNotFound(name string) {
	fmt.Println("code: dvx-17")
	fmt.Println("error: view '" + name + "' not found")
	fmt.Println("error_type: user error.")
	fmt.Println("message: list your views with 'devx view ls'")
	os.Exit(1)
}

func BunNotFound() {
	fmt.Println("code: dvx-18")
	fmt.Println("error: bun not found")
	fmt.Println("error_type: user error.")
	fmt.Println("message: install bun (https://bun.sh) and try again")
	os.Exit(1)
}

func CannotRunTool(err error) {
	fmt.Println("code: dvx-19")
	fmt.Println("error: error on run the frontend cli")
	fmt.Println("error_type: internal error.")
	fmt.Println("message: verify your network and bun installation. " + err.Error())
	os.Exit(1)
}

func MissingApp(app string) {
	fmt.Println("code: dvx-20")
	fmt.Println("error: this project has no " + app + " (cmd/" + app + ")")
	fmt.Println("error_type: user error.")
	fmt.Println("message: add it with 'devx app add " + app + "'")
	os.Exit(1)
}

func AppAlreadyExists(app string) {
	fmt.Println("code: dvx-22")
	fmt.Println("error: this project already has " + app + " (cmd/" + app + ")")
	fmt.Println("error_type: user error.")
	fmt.Println("message: see the project apps with 'devx app ls'")
	os.Exit(1)
}

func CommandAlreadyExists(name string) {
	fmt.Println("code: dvx-23")
	fmt.Println("error: command '" + name + "' already exists")
	fmt.Println("error_type: user error.")
	fmt.Println("message: list your commands with 'devx command ls'")
	os.Exit(1)
}

func CommandNotFound(name string) {
	fmt.Println("code: dvx-24")
	fmt.Println("error: command '" + name + "' not found")
	fmt.Println("error_type: user error.")
	fmt.Println("message: list your commands with 'devx command ls'")
	os.Exit(1)
}

func ReservedName(name string) {
	fmt.Println("code: dvx-25")
	fmt.Println("error: '" + name + "' is a reserved name")
	fmt.Println("error_type: user error.")
	fmt.Println("message: choose another name")
	os.Exit(1)
}

func ServiceNeedsApp(name, app string) {
	fmt.Println("code: dvx-26")
	fmt.Println("error: service '" + name + "' needs the " + app + " (cmd/" + app + ")")
	fmt.Println("error_type: user error.")
	fmt.Println("message: add it with 'devx app add " + app + "' or see 'devx service ls'")
	os.Exit(1)
}

func SudoRequired(action string) {
	fmt.Println("code: dvx-27")
	fmt.Println("error: " + action + " deletes a lot of code")
	fmt.Println("error_type: user error.")
	fmt.Println("message: pass --sudo if you really want it")
	os.Exit(1)
}

func LastApp(app string) {
	fmt.Println("code: dvx-28")
	fmt.Println("error: " + app + " is the only app of this project")
	fmt.Println("error_type: user error.")
	fmt.Println("message: a project needs at least one app; delete the folder to remove the project")
	os.Exit(1)
}

func NeedDBFlag() {
	fmt.Println("code: dvx-29")
	fmt.Println("error: this project has no database")
	fmt.Println("error_type: user error.")
	fmt.Println("message: not running interactively, pass --db postgres or --db memory")
	os.Exit(1)
}

func InvalidDB(db string) {
	fmt.Println("code: dvx-30")
	fmt.Println("error: unknown database '" + db + "'")
	fmt.Println("error_type: user error.")
	fmt.Println("message: try this options(postgres, memory)")
	os.Exit(1)
}

func CrudNotPostgres(name string) {
	fmt.Println("code: dvx-31")
	fmt.Println("error: crud '" + name + "' does not use postgres")
	fmt.Println("error_type: user error.")
	fmt.Println("message: access is set per postgres table, switch with 'devx crud db " + name + " postgres'")
	os.Exit(1)
}

func TableNotFound(table string) {
	fmt.Println("code: dvx-32")
	fmt.Println("error: table '" + table + "' not found in migrations")
	fmt.Println("error_type: user error.")
	fmt.Println("message: the crud table is not created by a migration (--no-migrations?), see 'devx crud ls'")
	os.Exit(1)
}

func UnknownOperation(op string) {
	fmt.Println("code: dvx-33")
	fmt.Println("error: unknown operation '" + op + "'")
	fmt.Println("error_type: user error.")
	fmt.Println("message: try this options(select, insert, update, delete)")
	os.Exit(1)
}

func ContainerIsBase(name string) {
	fmt.Println("code: dvx-34")
	fmt.Println("error: container '" + name + "' is part of the api, not a crud")
	fmt.Println("error_type: user error.")
	fmt.Println("message: it serves /ping, /docs and the openapi; remove it with 'devx crud rm server'")
	os.Exit(1)
}
