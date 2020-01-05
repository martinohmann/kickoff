package config

import (
	"fmt"

	"github.com/apex/log"
	"github.com/spf13/cobra"
	gitconfig "github.com/tcnksm/go-gitconfig"
)

const DefaultRepositoryHost = "github.com"

type RepositoryConfig struct {
	Host string `json:"host"`
	User string `json:"user"`
	Name string `json:"name"`
}

func NewDefaultRepositoryConfig() *RepositoryConfig {
	return &RepositoryConfig{
		Host: DefaultRepositoryHost,
	}
}

func (c *RepositoryConfig) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&c.User, "repository-user", c.User, "Repository user")
	cmd.Flags().StringVar(&c.Name, "repository-name", c.Name, "Repository name (defaults to the project name)")
	cmd.Flags().StringVar(&c.Host, "repository-host", c.Host, "Repository host")
}

func (c *RepositoryConfig) Complete(defaultName string) (err error) {
	if c.User == "" {
		c.User, err = gitconfig.Global("github.user")
		if err != nil {
			log.Warn("github.user not found in git config, set it to automatically populate repository user")
		}
	}

	if c.Name == "" {
		c.Name = defaultName
	}

	if c.Host == "" {
		c.Host = DefaultRepositoryHost
	}

	return nil
}

func (c *RepositoryConfig) Validate() error {
	if c.User == "" {
		return fmt.Errorf("--repository-user needs to be set as it could not be inferred")
	}

	return nil
}

// URL returns the repository url.
func (c *RepositoryConfig) URL() string {
	return fmt.Sprintf("https://%s/%s/%s", c.Host, c.User, c.Name)
}

// Package returns a string that can be used as a Golang package name for the
// project.
func (c *RepositoryConfig) Package() string {
	return fmt.Sprintf("%s/%s/%s", c.Host, c.User, c.Name)
}
