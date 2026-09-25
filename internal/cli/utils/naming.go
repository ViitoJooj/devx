package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var validName = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

type Names struct {
	Singular     string
	Plural       string
	Pascal       string
	PascalPlural string
	Camel        string
	CamelPlural  string
	Kebab        string
	KebabPlural  string
	Folder       string
	FolderPlural string
}

func NewNames(input string) (Names, error) {

	name := strings.ReplaceAll(strings.ToLower(input), "-", "_")
	if !validName.MatchString(name) {
		return Names{}, errors.New("invalid name")
	}

	plural := Pluralize(name)

	return Names{
		Singular:     name,
		Plural:       plural,
		Pascal:       pascal(name),
		PascalPlural: pascal(plural),
		Camel:        camel(name),
		CamelPlural:  camel(plural),
		Kebab:        strings.ReplaceAll(name, "_", "-"),
		KebabPlural:  strings.ReplaceAll(plural, "_", "-"),
		Folder:       strings.ReplaceAll(name, "_", ""),
		FolderPlural: strings.ReplaceAll(plural, "_", ""),
	}, nil
}

// NewEntityNames accepts the singular or the plural (user, users, component-definitions) and returns the singular names.
func NewEntityNames(input string) (Names, error) {

	name := strings.ReplaceAll(strings.ToLower(input), "-", "_")

	parts := strings.Split(name, "_")
	parts[len(parts)-1] = Singularize(parts[len(parts)-1])

	return NewNames(strings.Join(parts, "_"))
}

func Singularize(word string) string {
	switch {
	case strings.HasSuffix(word, "ies") && len(word) > 3:
		return word[:len(word)-3] + "y"
	case strings.HasSuffix(word, "sses"), strings.HasSuffix(word, "xes"), strings.HasSuffix(word, "zes"),
		strings.HasSuffix(word, "ches"), strings.HasSuffix(word, "shes"), strings.HasSuffix(word, "uses"):
		return word[:len(word)-2]
	case strings.HasSuffix(word, "ss"), strings.HasSuffix(word, "us"), strings.HasSuffix(word, "is"):
		return word
	case strings.HasSuffix(word, "s") && len(word) > 1:
		return word[:len(word)-1]
	default:
		return word
	}
}

func ValidName(input string) bool {
	return validName.MatchString(strings.ReplaceAll(input, "-", "_"))
}

func Pluralize(word string) string {
	switch {
	case strings.HasSuffix(word, "y") && len(word) > 1 && !strings.ContainsAny(word[len(word)-2:len(word)-1], "aeiou"):
		return word[:len(word)-1] + "ies"
	case strings.HasSuffix(word, "s"), strings.HasSuffix(word, "x"), strings.HasSuffix(word, "z"),
		strings.HasSuffix(word, "ch"), strings.HasSuffix(word, "sh"):
		return word + "es"
	default:
		return word + "s"
	}
}

func pascal(word string) string {
	var output strings.Builder
	for part := range strings.SplitSeq(word, "_") {
		if part == "" {
			continue
		}
		output.WriteString(strings.ToUpper(part[:1]) + part[1:])
	}
	return output.String()
}

func camel(word string) string {
	p := pascal(word)
	return strings.ToLower(p[:1]) + p[1:]
}

func GoDuration(d time.Duration) string {
	switch {
	case d%time.Hour == 0:
		return fmt.Sprintf("%d * time.Hour", d/time.Hour)
	case d%time.Minute == 0:
		return fmt.Sprintf("%d * time.Minute", d/time.Minute)
	case d%time.Second == 0:
		return fmt.Sprintf("%d * time.Second", d/time.Second)
	default:
		return fmt.Sprintf("%d * time.Millisecond", d/time.Millisecond)
	}
}
