package git

import (
	"fmt"

	"github.com/apex/log"
	gitconfig "github.com/tcnksm/go-gitconfig"
)

const DefaultHost = "github.com"

type Config struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	RepoName string `json:"-"`
}

func (c *Config) ApplyDefaults(defaultRepoName string) {
	var err error
	if c.User == "" {
		c.User, err = gitconfig.Global("github.user")
		if err != nil {
			log.Warn("github.user not found in git config, set it to automatically populate repository user")
		}
	}

	if c.RepoName == "" {
		c.RepoName = defaultRepoName
	}

	if c.Host == "" {
		c.Host = DefaultHost
	}
}

// URL returns the repository url.
func (c *Config) URL() string {
	return fmt.Sprintf("https://%s/%s/%s", c.Host, c.User, c.RepoName)
}

// GoPackagePath returns a string that can be used as a Golang package path for
// the project.
func (c *Config) GoPackagePath() string {
	return fmt.Sprintf("%s/%s/%s", c.Host, c.User, c.RepoName)
}
