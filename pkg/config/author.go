package config

import (
	"fmt"

	"github.com/apex/log"
	"github.com/spf13/cobra"
	gitconfig "github.com/tcnksm/go-gitconfig"
)

type AuthorConfig struct {
	Fullname string `json:"fullname"`
	Email    string `json:"email"`
}

func (c *AuthorConfig) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&c.Fullname, "author-fullname", c.Fullname, "Project author's fullname")
	cmd.Flags().StringVar(&c.Email, "author-email", c.Email, "Project author's e-mail")
}

func (c *AuthorConfig) Complete() (err error) {
	if c.Fullname == "" {
		c.Fullname, err = gitconfig.Global("user.name")
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

	return nil
}

func (c *AuthorConfig) String() string {
	if c.Email != "" {
		return fmt.Sprintf("%s <%s>", c.Fullname, c.Email)
	}

	return c.Fullname
}
