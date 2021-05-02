package kickoff

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/martinohmann/kickoff/internal/template"
	log "github.com/sirupsen/logrus"
	gitconfig "github.com/tcnksm/go-gitconfig"
)

// Config describes the schema for the local user-defined kickoff configuration
// file.
type Config struct {
	// Project holds default configuration values like default project host and
	// owner which are the same for new projects most of the time. Can be
	// overridden on project creation.
	Project ProjectConfig `json:"project,omitempty"`
	// Repositories holds a map of configured repositories to search for
	// available skeletons. Keys are the locally configured names for these
	// repositories.
	Repositories map[string]string `json:"repositories,omitempty"`
	// Values holds user-defined values that get merged on to of skeleton
	// values.
	Values template.Values `json:"values,omitempty"`
}

// DefaultConfig returns the default config.
func DefaultConfig() *Config {
	var config Config
	config.ApplyDefaults()
	return &config
}

// ApplyDefaults applies default values to the config.
func (c *Config) ApplyDefaults() {
	if c.Repositories == nil {
		c.Repositories = make(map[string]string)
	}

	if len(c.Repositories) == 0 {
		c.Repositories[DefaultRepositoryName] = DefaultRepositoryURL
	}

	if c.Values == nil {
		c.Values = make(template.Values)
	}

	c.Project.ApplyDefaults()
}

// Validate implements the Validator interface.
func (c *Config) Validate() error {
	for name, repoURL := range c.Repositories {
		if name == "" {
			return newRepositoryRefError("repository name must not be empty")
		}

		if !repoNameRegexp.MatchString(name) {
			return newRepositoryRefError("repository name %q does not match pattern: %s", name, repoNameRegexp)
		}

		if repoURL == "" {
			return newRepositoryRefError("repository URL must not be empty")
		}

		if _, err := url.Parse(repoURL); err != nil {
			return newRepositoryRefError("invalid URL: %w", err)
		}
	}

	return c.Project.Validate()
}

// ProjectConfig contains project specific configuration like git host, owner and
// project name.
type ProjectConfig struct {
	// Host holds the default git host's domain, e.g. 'github.com'.
	Host string `json:"host,omitempty"`
	// Owner holds the default repository owner name.
	Owner string `json:"owner,omitempty"`
	// License holds the name of the default open source license.
	License string `json:"license,omitempty"`
	// Gitignore holds a comma-separated list of gitignore templates, e.g.
	// 'go,hugo'.
	Gitignore string `json:"gitignore,omitempty"`
}

// ApplyDefaults applies defaults to unset fields. If the Owner field is empty
// ApplyDefaults will attempt to fill it with the git config values of
// github.user or user.name if exists.
func (p *ProjectConfig) ApplyDefaults() {
	if p.Host == "" {
		p.Host = DefaultProjectHost
	}

	if p.Owner == "" {
		p.Owner = detectDefaultProjectOwner()
	}
}

// Validate implements the Validator interface.
func (c *ProjectConfig) Validate() error {
	if c.Host != "" {
		if _, err := url.Parse(c.Host); err != nil {
			return newProjectConfigError("invalid Host: %w", err)
		}
	}

	return nil
}

// Load loads the config from path and returns it.
func LoadConfig(path string) (*Config, error) {
	var config Config

	err := Load(path, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	err = config.Validate()
	if err != nil {
		return nil, err
	}

	config.ApplyDefaults()

	return &config, nil
}

// SaveConfig saves config to path.
func SaveConfig(path string, config *Config) error {
	err := config.Validate()
	if err != nil {
		return err
	}

	return Save(path, config)
}

// Load loads a file from path into v. Returns an error if reading the file
// fails. Does not perform any defaulting or validation.
func Load(path string, v interface{}) error {
	log.WithField("path", path).Debug("loading file")

	buf, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(buf, &v)
}

// Save saves v to path.
func Save(path string, v interface{}) error {
	log.WithField("path", path).Debug("saving file")

	buf, err := yaml.Marshal(v)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return err
	}

	return os.WriteFile(path, buf, 0644)
}

func detectDefaultProjectOwner() string {
	configKeys := []string{"github.user", "user.name"}

	owner := lookupGitConfig(configKeys)
	if owner == "" {
		log.Infof("could not infer project owner from git config, none of these keys found: %s", strings.Join(configKeys, ", "))
	}

	return owner
}

var gitconfigFn = gitconfig.Global // for mocking in tests

func lookupGitConfig(configKeys []string) string {
	for _, key := range configKeys {
		if val, err := gitconfigFn(key); err == nil {
			return val
		}
	}

	return ""
}
