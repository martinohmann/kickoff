package project

import (
	"fmt"

	"github.com/apex/log"
	gitconfig "github.com/tcnksm/go-gitconfig"
)

type Config struct {
	Name   string `json:"-"`
	Author string `json:"author"`
	Email  string `json:"email"`
}

// ApplyDefaults applies defaults to unset fields.
func (c *Config) ApplyDefaults(defaultProjectName string) {
	if c.Name == "" {
		c.Name = defaultProjectName
	}

	var err error
	if c.Author == "" {
		c.Author, err = gitconfig.Global("user.name")
		if err != nil {
			log.Warn("user.name not found in git config, set it to automatically populate author fullname")
		}
	}

	if c.Email == "" {
		c.Email, err = gitconfig.Global("user.email")
		if err != nil {
			log.Warn("user.email not found in git config, set it to automatically populate author email")
		}
	}
}

func (c *Config) AuthorString() string {
	if c.Email != "" {
		return fmt.Sprintf("%s <%s>", c.Author, c.Email)
	}

	return c.Author
}
