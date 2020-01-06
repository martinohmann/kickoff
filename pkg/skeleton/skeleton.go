package skeleton

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/imdario/mergo"
)

const (
	// ConfigFile is the name of the skeleton's config file.
	ConfigFile = ".kickoff.yaml"
)

// Info holds information about a skeleton.
type Info struct {
	Name string
	Path string
}

// Config loads the skeleton config for the info.
func (i *Info) Config() (*Config, error) {
	return LoadConfig(filepath.Join(i.Path, ConfigFile))
}

// Walk recursively walks all files and directories of the skeleton. This
// behaves exactly as filepath.Walk, except that it will ignore the skeleton's
// ConfigFile.
func (i *Info) Walk(walkFn filepath.WalkFunc) error {
	return filepath.Walk(i.Path, func(path string, info os.FileInfo, err error) error {
		if info.Name() == ConfigFile {
			// ignore skeleton config file
			return err
		}

		return walkFn(path, info, err)
	})
}

// Config defines the structure of a ConfigFile.
type Config struct {
	License string                 `json:"license"`
	Values  map[string]interface{} `json:"values"`
}

// Merge merges other on top of c. Non-zero fields in other will override the
// same fields in c.
func (c *Config) Merge(other *Config) error {
	return mergo.Merge(c, other, mergo.WithOverride)
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
