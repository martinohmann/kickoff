// Package template provides tools to render template strings and template
// files.
package template

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

// RenderFile renders a template file with data.
func RenderFile(file string, data interface{}) (string, error) {
	name := filepath.Base(file)

	tpl, err := newTemplate(name).ParseFiles(file)
	if err != nil {
		return "", fmt.Errorf("failed to prepare template: %v", err)
	}

	return execute(tpl, data)
}

// Render renders template text with data.
func RenderText(templateText string, data interface{}) (string, error) {
	tpl, err := newTemplate("").Parse(templateText)
	if err != nil {
		return "", fmt.Errorf("failed to prepare template: %v", err)
	}

	return execute(tpl, data)
}

func newTemplate(name string) *template.Template {
	return template.New(name).
		Option("missingkey=error").
		Funcs(sprig.TxtFuncMap()).
		Funcs(funcMap)
}

func execute(tpl *template.Template, data interface{}) (string, error) {
	var buf bytes.Buffer

	err := tpl.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("failed to render template: %v", err)
	}

	return buf.String(), nil
}
