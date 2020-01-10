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
	localConfigDir = configdir.LocalConfig("kickoff")

	DefaultConfigPath            = filepath.Join(localConfigDir, "config.yaml")
	DefaultSkeletonRepositoryURL = filepath.Join(localConfigDir, "repository")
)

const (
	DefaultGitHost = "github.com"
	DefaultLicense = "none"

	// SkeletonConfigFile is the name of the skeleton's config file.
	SkeletonConfigFile = ".kickoff.yaml"
)

type Config struct {
	License   string          `json:"license"`
	Project   Project         `json:"project"`
	Git       Git             `json:"git"`
	Skeletons Skeletons       `json:"skeletons"`
	Values    template.Values `json:"values"`
}

func (c *Config) ApplyDefaults(defaultProjectName string) {
	if c.License == "" {
		c.License = DefaultLicense
	}

	c.Project.ApplyDefaults(defaultProjectName)
	c.Git.ApplyDefaults(c.Project.Name)
	c.Skeletons.ApplyDefaults()
}

// MergeFromFile loads the config from path and merges it into c. Returns any
// errors that may occur during loading or merging. Non-zero fields in c will
// not be overridden if present in the file at path.
func (c *Config) MergeFromFile(path string) error {
	config, err := Load(path)
	if err != nil {
		return err
	}

	return mergo.Merge(c, config)
}

func (c *Config) HasLicense() bool {
	return c.License != "" && c.License != "none"
}

type Project struct {
	Author string `json:"author"`
	Email  string `json:"email"`
	Name   string `json:"-"`
}

// ApplyDefaults applies defaults to unset fields.
func (p *Project) ApplyDefaults(defaultName string) {
	if p.Name == "" {
		p.Name = defaultName
	}

	var err error
	if p.Author == "" {
		p.Author, err = gitconfig.Global("user.name")
		if err != nil {
			log.Warn("user.name not found in git config, set it to automatically populate author fullname")
		}
	}

	if p.Email == "" {
		p.Email, err = gitconfig.Global("user.email")
		if err != nil {
			log.Warn("user.email not found in git config, set it to automatically populate author email")
		}
	}
}

func (p *Project) AuthorString() string {
	if p.Email != "" {
		return fmt.Sprintf("%s <%s>", p.Author, p.Email)
	}

	return p.Author
}

type Git struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	RepoName string `json:"-"`
}

func (g *Git) ApplyDefaults(defaultRepoName string) {
	var err error
	if g.User == "" {
		g.User, err = gitconfig.Global("github.user")
		if err != nil {
			log.Warn("github.user not found in git config, set it to automatically populate repository user")
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

type Skeletons struct {
	RepositoryURL string `json:"repositoryURL"`
}

func (s *Skeletons) ApplyDefaults() {
	if s.RepositoryURL == "" {
		s.RepositoryURL = DefaultSkeletonRepositoryURL
	}
}

type Skeleton struct {
	Values template.Values `json:"values"`
}

// Load loads the config from path and returns it.
func Load(path string) (Config, error) {
	var config Config

	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(buf, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

// LoadSkeleton loads the skeleton config from path and returns it.
func LoadSkeleton(path string) (Skeleton, error) {
	var config Skeleton

	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(buf, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
