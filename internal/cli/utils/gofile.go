package utils

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

type GoFile struct {
	Path  string
	Src   []byte
	Fset  *token.FileSet
	File  *ast.File
	Lines []string
}

// Anchor returns the line (1-based) where a snippet is inserted and the indentation it receives.
type Anchor func(g *GoFile) (line int, indent string, err error)

func LoadGo(path string) (*GoFile, error) {

	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	return &GoFile{
		Path:  path,
		Src:   src,
		Fset:  fset,
		File:  file,
		Lines: strings.Split(string(src), "\n"),
	}, nil
}

// InsertGo writes snippet at the anchor, formats the file and skips snippets already present.
func InsertGo(path string, anchor Anchor, snippet string) error {

	g, err := LoadGo(path)
	if err != nil {
		return err
	}

	if alreadyInjected(string(g.Src), snippet) {
		return nil
	}

	line, indent, err := anchor(g)
	if err != nil {
		return err
	}

	var block []string
	for snippetLine := range strings.SplitSeq(strings.TrimSuffix(snippet, "\n"), "\n") {
		if snippetLine == "" {
			block = append(block, "")
			continue
		}
		block = append(block, indent+snippetLine)
	}

	index := line - 1
	lines := append(g.Lines[:index:index], append(block, g.Lines[index:]...)...)

	return writeFormatted(path, strings.Join(lines, "\n"))
}

func (g *GoFile) Line(pos token.Pos) int {
	return g.Fset.Position(pos).Line
}

func (g *GoFile) Text(node ast.Node) string {
	return string(g.Src[g.Fset.Position(node.Pos()).Offset:g.Fset.Position(node.End()).Offset])
}

func (g *GoFile) IndentOf(line int) string {
	text := g.Lines[line-1]
	return text[:len(text)-len(strings.TrimLeft(text, " \t"))]
}

func (g *GoFile) Func(name string) *ast.FuncDecl {
	for _, decl := range g.File.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == name && fn.Recv == nil {
			return fn
		}
	}
	return nil
}

// Stmt finds the first statement of a function whose source starts with prefix.
func (g *GoFile) Stmt(funcName, prefix string) ast.Stmt {

	fn := g.Func(funcName)
	if fn == nil || fn.Body == nil {
		return nil
	}

	for _, stmt := range fn.Body.List {
		if strings.HasPrefix(g.Text(stmt), prefix) {
			return stmt
		}
	}
	return nil
}

// LastStmt finds the last statement of a function whose source starts with any prefix.
func (g *GoFile) LastStmt(funcName string, match func(text string) bool) ast.Stmt {

	fn := g.Func(funcName)
	if fn == nil || fn.Body == nil {
		return nil
	}

	var last ast.Stmt
	for _, stmt := range fn.Body.List {
		if match(g.Text(stmt)) {
			last = stmt
		}
	}
	return last
}

// Call finds the first call whose function is written as fun (ex: workers.New).
func (g *GoFile) Call(fun string) *ast.CallExpr {

	var found *ast.CallExpr
	ast.Inspect(g.File, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if ok && found == nil && g.Text(call.Fun) == fun {
			found = call
		}
		return found == nil
	})
	return found
}

func (g *GoFile) Struct(name string) *ast.StructType {

	var found *ast.StructType
	ast.Inspect(g.File, func(node ast.Node) bool {
		spec, ok := node.(*ast.TypeSpec)
		if ok && spec.Name.Name == name {
			found, _ = spec.Type.(*ast.StructType)
		}
		return found == nil
	})
	return found
}

// Composite finds the first composite literal of the given type (ex: Cfg{...}).
func (g *GoFile) Composite(typeName string) *ast.CompositeLit {

	var found *ast.CompositeLit
	ast.Inspect(g.File, func(node ast.Node) bool {
		lit, ok := node.(*ast.CompositeLit)
		if ok && found == nil && lit.Type != nil && g.Text(lit.Type) == typeName {
			found = lit
		}
		return found == nil
	})
	return found
}

func (g *GoFile) Imports() *ast.GenDecl {
	for _, decl := range g.File.Decls {
		if gen, ok := decl.(*ast.GenDecl); ok && gen.Tok == token.IMPORT {
			return gen
		}
	}
	return nil
}

// ImportAnchor inserts inside the import (...) block.
func ImportAnchor(g *GoFile) (int, string, error) {

	imports := g.Imports()
	if imports == nil || !imports.Rparen.IsValid() {
		return 0, "", errors.New(g.Path + ": import block with parentheses not found")
	}

	return g.Line(imports.Rparen), "\t", nil
}

// FirstStmt finds the first statement of a function whose source matches.
func (g *GoFile) FirstStmt(funcName string, match func(text string) bool) ast.Stmt {

	fn := g.Func(funcName)
	if fn == nil || fn.Body == nil {
		return nil
	}

	for _, stmt := range fn.Body.List {
		if match(g.Text(stmt)) {
			return stmt
		}
	}
	return nil
}

// RemoveGoStmt deletes the first statement of a function whose source starts with prefix.
func RemoveGoStmt(path, funcName, prefix string) error {

	g, err := LoadGo(path)
	if err != nil {
		return err
	}

	stmt := g.Stmt(funcName, prefix)
	if stmt == nil {
		return nil
	}

	start, end := g.Line(stmt.Pos())-1, g.Line(stmt.End())
	lines := append(g.Lines[:start:start], g.Lines[end:]...)
	lines = dropDoubleBlank(lines, start)

	if start < len(lines) && start > 0 && strings.TrimSpace(lines[start]) == "" && strings.HasSuffix(strings.TrimSpace(lines[start-1]), "{") {
		lines = append(lines[:start], lines[start+1:]...)
	}

	return writeFormatted(path, strings.Join(lines, "\n"))
}
