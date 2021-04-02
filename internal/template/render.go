// Package template provides tools to render template strings and template
// files.
package template

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
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

// RenderReader renders a template obtained from a reader with data.
func RenderReader(r io.Reader, data interface{}) (string, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}

	return Render(string(buf), data)
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
