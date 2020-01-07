package kickoff

import (
	"io/ioutil"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/imdario/mergo"
	"github.com/kirsle/configdir"
	"github.com/martinohmann/kickoff/pkg/git"
	"github.com/martinohmann/kickoff/pkg/project"
	"github.com/martinohmann/kickoff/pkg/repo"
	"github.com/martinohmann/kickoff/pkg/template"
)

const (
	DefaultLicense = "none"
)

var (
	localConfigDir    = configdir.LocalConfig("kickoff")
	DefaultConfigPath = filepath.Join(localConfigDir, "config.yaml")
)

type Config struct {
	License string          `json:"license"`
	Project project.Config  `json:"project"`
	Git     git.Config      `json:"git"`
	Repo    repo.Config     `json:"repository"`
	Values  template.Values `json:"values"`
}

func (c *Config) ApplyDefaults(defaultProjectName string) {
	if c.License == "" {
		c.License = DefaultLicense
	}

	c.Project.ApplyDefaults(defaultProjectName)
	c.Git.ApplyDefaults(c.Project.Name)
	c.Repo.ApplyDefaults()
}

// MergeFromFile loads the config from path and merges it into c. Returns any
// errors that may occur during loading or merging. Non-zero fields in c will
// not be overridden if present in the file at path.
func (c *Config) MergeFromFile(path string) error {
	config, err := LoadConfig(path)
	if err != nil {
		return err
	}

	return mergo.Merge(c, config)
}

// LoadConfig loads the config from path and returns it.
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
