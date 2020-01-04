package template

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

// Render renders a template file with data.
func Render(file string, data interface{}) ([]byte, error) {
	name := filepath.Base(file)

	tpl, err := template.New(name).
		Option("missingkey=error").
		Funcs(sprig.TxtFuncMap()).
		ParseFiles(file)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare template: %v", err)
	}

	var buf bytes.Buffer

	err = tpl.Execute(&buf, data)
	if err != nil {
		return nil, fmt.Errorf("failed to render template: %v", err)
	}

	return buf.Bytes(), nil
}
