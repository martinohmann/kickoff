// Package config provides configuration for kickoff.
package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/ghodss/yaml"
	"github.com/imdario/mergo"
	"github.com/kirsle/configdir"
	"github.com/martinohmann/kickoff/pkg/template"
	gitconfig "github.com/tcnksm/go-gitconfig"
)

var (
	// LocalConfigDir points to the user's local configuration dir which is
	// platform specific.
	LocalConfigDir = configdir.LocalConfig("kickoff")

	// DefaultConfigPath holds the default kickof config path in the user's
	// local config directory.
	DefaultConfigPath = filepath.Join(LocalConfigDir, "config.yaml")

	// DefaultRepositoryURL is the url of the default skeleton repository if
	// the user did not configure anything else.
	DefaultRepositoryURL = "https://github.com/martinohmann/kickoff-skeletons"
)

const (
	// DefaultProjectHost denotes the default git host that is passed to
	// templates so that project related urls can be rendered in files like
	// READMEs.
	DefaultProjectHost = "github.com"

	// DefaultRepositoryName is the name of the default skeleton repository.
	DefaultRepositoryName = "default"

	// DefaultSkeletonName is the name of the default skeleton in a repository.
	DefaultSkeletonName = "default"

	// NoLicense means that no license file will be generated for a new
	// project.
	NoLicense = "none"

	// NoGitignore means that no .gitignore file will be generated for a new
	// project.
	NoGitignore = "none"
)

// Config is the type for user-defined configuration.
type Config struct {
	Project      Project           `json:"project"`
	Repositories map[string]string `json:"repositories"`
	Values       template.Values   `json:"values"`
}

// ApplyDefaults applies default values to the config.
func (c *Config) ApplyDefaults() {
	if c.Repositories == nil {
		c.Repositories = make(map[string]string)
	}

	if len(c.Repositories) == 0 {
		c.Repositories[DefaultRepositoryName] = DefaultRepositoryURL
	}

	c.Project.ApplyDefaults()

	if c.Values == nil {
		c.Values = template.Values{}
	}
}

// MergeFromFile loads the config from path and merges it into c. Returns any
// errors that may occur during loading or merging. Non-zero fields in c will
// not be overridden if present in the file at path.
func (c *Config) MergeFromFile(path string) error {
	config, err := Load(path)
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	return mergo.Merge(c, config)
}

// Project contains project specific configuration like git host, owner and
// project name.
type Project struct {
	Host      string `json:"host"`
	Owner     string `json:"owner"`
	Name      string `json:"-"`
	License   string `json:"license"`
	Gitignore string `json:"gitignore"`
}

// ApplyDefaults applies defaults to unset fields. If the Owner field is empty
// ApplyDefaults will attempt to fill it with the git config values of
// github.user or user.name if exists.
func (p *Project) ApplyDefaults() {
	if p.Host == "" {
		p.Host = DefaultProjectHost
	}

	if p.Owner == "" {
		p.Owner = detectProjectOwner()
	}

	if p.License == "" {
		p.License = NoLicense
	}

	if p.Gitignore == "" {
		p.Gitignore = NoGitignore
	}
}

// HasLicense returns true if an open source license is specified in the
// project config. If true, the project creator will write the text of the
// provided license into the LICENSE file in the project's output directory.
func (p *Project) HasLicense() bool {
	return p.License != "" && p.License != NoLicense
}

// HasGitignore returns true if a gitignore template is specified in the
// project config. If true, the project creator will write the gitignore
// template into the .gitignore file in the project's output directory.
func (p *Project) HasGitignore() bool {
	return p.Gitignore != "" && p.Gitignore != NoGitignore
}

// URL returns the repository url.
func (p *Project) URL() string {
	return fmt.Sprintf("https://%s/%s/%s", p.Host, p.Owner, p.Name)
}

// GoPackagePath returns a string that can be used as a Golang package path for
// the project.
func (p *Project) GoPackagePath() string {
	return fmt.Sprintf("%s/%s/%s", p.Host, p.Owner, p.Name)
}

// Load loads the config from path and returns it.
func Load(path string) (Config, error) {
	var config Config

	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(buf, &config)

	return config, err
}

// Save saves config to path.
func Save(config *Config, path string) error {
	buf, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, buf, 0644)
}

func getGitConfigKey(configKeys []string) string {
	for _, key := range configKeys {
		val, err := gitconfig.Global(key)
		if err == nil {
			return val
		}
	}

	return ""
}

func detectProjectOwner() string {
	configKeys := []string{"github.user", "user.name"}

	owner := getGitConfigKey(configKeys)
	if owner == "" {
		log.Debugf("could not infer project owner from git config, none of these keys found: ", strings.Join(configKeys, ", "))
	}

	return owner
}
