package cmdutil

import (
	"encoding/json"
	"io"

	"github.com/ghodss/yaml"
)

// RenderJSON converts v to JSON and writes it to w.
func RenderJSON(w io.Writer, v interface{}) error {
	buf, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	_, err = w.Write(buf)
	return err
}

// RenderYAML converts v to YAML and writes it to w.
func RenderYAML(w io.Writer, v interface{}) error {
	buf, err := yaml.Marshal(v)
	if err != nil {
		return err
	}

	_, err = w.Write(buf)
	return err
}
