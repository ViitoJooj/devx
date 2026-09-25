package utils

import (
	"bytes"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/ViitoJooj/devx/internal/templates"
)

const tmplExt = ".tmpl"

type RenderInput struct {
	Source       string
	Destination  string
	Data         any
	Placeholders map[string]string
	SkipExisting bool
	Exclude      []string
}

func RenderDir(input RenderInput) ([]string, error) {

	var created []string

	err := fs.WalkDir(templates.FS, input.Source, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		rel := strings.TrimPrefix(path, input.Source+"/")
		if excluded(rel, input.Exclude) {
			return nil
		}

		for key, value := range input.Placeholders {
			rel = strings.ReplaceAll(rel, key, value)
		}
		rel = strings.TrimSuffix(rel, tmplExt)

		content, err := fs.ReadFile(templates.FS, path)
		if err != nil {
			return err
		}

		output, err := Render(path, string(content), input.Data)
		if err != nil {
			return err
		}

		dest := filepath.Join(input.Destination, rel)
		if _, err := os.Stat(dest); err == nil && input.SkipExisting {
			return nil
		}

		if err := WriteNewFile(dest, output); err != nil {
			return err
		}

		created = append(created, rel)
		return nil
	})

	return created, err
}

func Render(name, content string, data any) (string, error) {

	tmpl, err := template.New(name).Delims("{%", "%}").Parse(content)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	if strings.HasSuffix(strings.TrimSuffix(name, tmplExt), ".go") {
		formatted, err := format.Source(buf.Bytes())
		if err != nil {
			return "", fmt.Errorf("%s: %w", name, err)
		}
		return string(formatted), nil
	}

	return buf.String(), nil
}

func WriteNewFile(path, content string) error {

	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists", path)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(content), 0o644)
}

func excluded(rel string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(rel, prefix) {
			return true
		}
	}
	return false
}
