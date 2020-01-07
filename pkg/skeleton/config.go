package skeleton

import (
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/martinohmann/kickoff/pkg/template"
)

// Config defines the structure of a ConfigFile.
type Config struct {
	Values template.Values `json:"values"`
}

// LoadConfig loads the skeleton config from path and returns it.
func LoadConfig(path string) (*Config, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config

	err = yaml.Unmarshal(buf, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
