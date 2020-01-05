package template

import (
	"regexp"
	"strings"
	"text/template"

	"github.com/ghodss/yaml"
)

var nonLetterDigitRegexp = regexp.MustCompile("[^a-zA-Z0-9]+")

var funcMap = template.FuncMap{
	"goPackageName": goPackageName,
	"toYAML":        toYAML,
	"mustToYAML":    mustToYAML,

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

func goPackageName(s string) string {
	parts := strings.Split(s, "/")
	last := parts[len(parts)-1]
	s = nonLetterDigitRegexp.ReplaceAllString(last, "")
	return strings.ToLower(s)
}
