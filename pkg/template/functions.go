package template

import (
	"text/template"

	"github.com/ghodss/yaml"
)

var funcMap = template.FuncMap{
	"toYAML":     toYAML,
	"mustToYAML": mustToYAML,

	// For compatibility with the naming helm users are used to.
	"toYaml":     toYAML,
	"mustToYaml": mustToYAML,
}

func toYAML(data interface{}) string {
	s, _ := mustToYAML(data)
	return s
}

func mustToYAML(data interface{}) (string, error) {
	buf, err := yaml.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}
