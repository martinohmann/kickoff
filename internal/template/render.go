// Package template provides tools to render template strings and template
// files.
package template

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

// Render renders template text with data.
func Render(templateText string, data interface{}) (string, error) {
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

// Renderer can render multiple templates with the same values.
type Renderer struct {
	values Values
}

// NewRenderer creates a new *Renderer which injects values in every template
// it renders.
func NewRenderer(values Values) *Renderer {
	return &Renderer{values}
}

// Render renders templateText.
func (r *Renderer) Render(templateText string) (string, error) {
	return Render(templateText, r.values)
}
