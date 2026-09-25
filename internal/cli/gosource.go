package cli

import (
	"errors"
	"go/ast"
	"strings"

	"github.com/ViitoJooj/devx/internal/cli/utils"
)

func anchorError(g *utils.GoFile, what string) error {
	return errors.New(g.Path + ": " + what + " not found (file changed by hand?)")
}

func addImport(path, spec string) error {
	return utils.InsertGo(path, utils.ImportAnchor, spec)
}

const apiRun = "if err := api.Run("

func isContainerInit(text string) bool {
	return strings.HasPrefix(text, "if err := ") && strings.Contains(text, ".Init(ctx, api")
}

func stmtAnchor(g *utils.GoFile, stmt ast.Stmt, what string) (int, string, error) {
	if stmt == nil {
		return 0, "", anchorError(g, what)
	}

	line := g.Line(stmt.Pos())
	return line, g.IndentOf(line), nil
}

func middlewareAnchor(g *utils.GoFile) (int, string, error) {
	return stmtAnchor(g, g.Stmt("main", "defer cancel()"), "defer cancel()")
}

func connAnchor(g *utils.GoFile) (int, string, error) {

	stmt := g.Stmt("main", "workers.New(")
	if stmt == nil {
		stmt = firstContainerOrRun(g)
	}

	return stmtAnchor(g, stmt, "api.Run")
}

func workersAnchor(g *utils.GoFile) (int, string, error) {
	return stmtAnchor(g, firstContainerOrRun(g), "api.Run")
}

func containerAnchor(g *utils.GoFile) (int, string, error) {

	last := g.LastStmt("main", isContainerInit)
	if last == nil {
		return stmtAnchor(g, g.Stmt("main", apiRun), "api.Run")
	}

	return g.Line(last.End()) + 1, g.IndentOf(g.Line(last.Pos())), nil
}

// firstContainerOrRun is where code that runs before the containers goes; without containers (server removed), before api.Run.
func firstContainerOrRun(g *utils.GoFile) ast.Stmt {
	if stmt := g.FirstStmt("main", isContainerInit); stmt != nil {
		return stmt
	}
	return g.Stmt("main", apiRun)
}

func beforeStmtAnchor(fn, prefix string) utils.Anchor {
	return func(g *utils.GoFile) (int, string, error) {
		return stmtAnchor(g, g.Stmt(fn, prefix), prefix)
	}
}

func callArgAnchor(fun string) utils.Anchor {
	return func(g *utils.GoFile) (int, string, error) {

		call := g.Call(fun)
		if call == nil {
			return 0, "", anchorError(g, fun)
		}

		line := g.Line(call.Rparen)
		return line, g.IndentOf(line) + "\t", nil
	}
}

func cfgFieldAnchor(g *utils.GoFile) (int, string, error) {

	cfg := g.Struct("Cfg")
	if cfg == nil {
		return 0, "", anchorError(g, "type Cfg struct")
	}

	line := g.Line(cfg.Fields.Closing)
	return line, g.IndentOf(line) + "\t", nil
}

func cfgTypeAnchor(g *utils.GoFile) (int, string, error) {

	fn := g.Func("NewDotenv")
	if fn == nil {
		return 0, "", anchorError(g, "func NewDotenv")
	}

	return g.Line(fn.Pos()), "", nil
}

func cfgValueAnchor(g *utils.GoFile) (int, string, error) {

	lit := g.Composite("Cfg")
	if lit == nil {
		return 0, "", anchorError(g, "Cfg{...}")
	}

	line := g.Line(lit.Rbrace)
	return line, g.IndentOf(line) + "\t", nil
}

var cfgValidateAnchor = beforeStmtAnchor("validate", "return dotenvNullValue")
