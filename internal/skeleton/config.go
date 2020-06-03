package skeleton

import (
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/martinohmann/kickoff/internal/template"
)

// ConfigFileName is the name of the skeleton's config file.
const ConfigFileName = ".kickoff.yaml"

// Config describes the schema of a skeleton's .kickoff.yaml.
type Config struct {
	Description string          `json:"description,omitempty"`
	Parent      *Reference      `json:"parent,omitempty"`
	Values      template.Values `json:"values"`
}

// Reference is a reference to a skeleton in a specific repository.
type Reference struct {
	RepositoryURL string `json:"repositoryURL,omitempty"`
	SkeletonName  string `json:"skeletonName"`
}

// LoadConfig loads the skeleton config from path and returns it.
func LoadConfig(path string) (Config, error) {
	var config Config

	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(buf, &config)

	return config, err
}
