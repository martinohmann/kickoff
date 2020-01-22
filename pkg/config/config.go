// Package config provides configuration for kickoff and kickoff skeletons.
package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

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

	// DefaultRepositoryURL is the default lookup path for the user's
	// local skeleton directory.
	DefaultRepositoryURL = filepath.Join(LocalConfigDir, "repository")

	// DefaultRepositoryName is the name of the default skeleton repository.
	DefaultRepositoryName = "default"

	// DefaultSkeletonName is the name of the default skeleton in a repository.
	DefaultSkeletonName = "default"
)

const (
	// DefaultGitHost denotes the default git host that is passed to templates so
	// that repository related urls can be rendered in files like READMEs.
	DefaultGitHost = "github.com"

	// NoLicense means that no license file will be generated for a new
	// project.
	NoLicense = "none"

	// NoGitignore means that no .gitignore file will be generated for a new
	// project.
	NoGitignore = "none"
)

// Config is the type for user-defined configuration.
type Config struct {
	License      string            `json:"license"`
	Gitignore    string            `json:"gitignore"`
	Project      Project           `json:"project"`
	Git          Git               `json:"git"`
	Repositories map[string]string `json:"repositories"`
	Values       template.Values   `json:"values"`
}

// ApplyDefaults applies default values to the config. The defaultProjectName
// variable will be used to set the project name and the git repository name
// (if they are unset).
func (c *Config) ApplyDefaults(defaultProjectName string) {
	if c.License == "" {
		c.License = NoLicense
	}

	if c.Gitignore == "" {
		c.Gitignore = NoGitignore
	}

	if c.Repositories == nil {
		c.Repositories = make(map[string]string)
	}

	_, ok := c.Repositories[DefaultRepositoryName]
	if !ok {
		c.Repositories[DefaultRepositoryName] = DefaultRepositoryURL
	}

	c.Project.ApplyDefaults(defaultProjectName)
	c.Git.ApplyDefaults(c.Project.Name)

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

// HasLicense returns true if an open source license is specified in the
// config. If true, the project creator will write the text of the provided
// license into the LICENSE file in the project's output directory.
func (c *Config) HasLicense() bool {
	return c.License != "" && c.License != NoLicense
}

// HasGitignore returns true if a gitignore template is specified in the
// config. If true, the project creator will write the gitignore template into
// the .gitignore file in the project's output directory.
func (c *Config) HasGitignore() bool {
	return c.Gitignore != "" && c.Gitignore != NoGitignore
}

// Project contains project specific configuration like author, email address
// and project name.
type Project struct {
	Author string `json:"author"`
	Email  string `json:"email"`
	Name   string `json:"-"`
}

// ApplyDefaults applies defaults to unset fields. If the Author and Email
// fields are empty ApplyDefaults will attempt to fill them with the git config
// values of user.name and user.email if they exist.
func (p *Project) ApplyDefaults(defaultName string) {
	if p.Name == "" {
		p.Name = defaultName
	}

	var err error
	if p.Author == "" {
		p.Author, err = gitconfig.Global("user.name")
		if err != nil {
			log.Debug("user.name not found in git config, set it to automatically populate author fullname")
		}
	}

	if p.Email == "" {
		p.Email, err = gitconfig.Global("user.email")
		if err != nil {
			log.Debug("user.email not found in git config, set it to automatically populate author email")
		}
	}
}

// AuthorString a string that can be used in licenses. If an email address is
// configured, this will look like `Author <Email>`. `Author` otherwise.
func (p *Project) AuthorString() string {
	if p.Email != "" {
		return fmt.Sprintf("%s <%s>", p.Author, p.Email)
	}

	return p.Author
}

// Git holds information about the project repository. These values are
// currently only forwarded to templates so that users can template links
// related to their project in README files and the like.
type Git struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	RepoName string `json:"-"`
}

// ApplyDefaults applies git configuration defaults. If the User field is empty
// ApplyDefaults will attempt to fill it with the git config values of
// github.user if it exist.
func (g *Git) ApplyDefaults(defaultRepoName string) {
	var err error
	if g.User == "" {
		g.User, err = gitconfig.Global("github.user")
		if err != nil {
			log.Debug("github.user not found in git config, set it to automatically populate repository user")
		}
	}

	if g.RepoName == "" {
		g.RepoName = defaultRepoName
	}

	if g.Host == "" {
		g.Host = DefaultGitHost
	}
}

// URL returns the repository url.
func (g *Git) URL() string {
	return fmt.Sprintf("https://%s/%s/%s", g.Host, g.User, g.RepoName)
}

// GoPackagePath returns a string that can be used as a Golang package path for
// the project.
func (g *Git) GoPackagePath() string {
	return fmt.Sprintf("%s/%s/%s", g.Host, g.User, g.RepoName)
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
