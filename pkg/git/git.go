package git

import (
	"github.com/apex/log"
	gitconfig "github.com/tcnksm/go-gitconfig"
	git "gopkg.in/src-d/go-git.v4"
)

type Config struct {
	Username   string
	Email      string
	GitHubUser string
}

func GlobalConfig() *Config {
	username, err := gitconfig.Global("user.name")
	if err != nil {
		log.Warn("user.name not found in git config, set it to automatically populate author fullname")
	}

	email, err := gitconfig.Global("user.email")
	if err != nil {
		log.Warn("user.email not found in git config, set it to automatically populate author email")
	}

	githubUser, err := gitconfig.Global("github.user")
	if err != nil {
		log.Warn("github.user not found in git config, set it to automatically populate repository user")
	}

	return &Config{
		Username:   username,
		Email:      email,
		GitHubUser: githubUser,
	}
}

func IsRepository(path string) bool {
	if _, err := git.PlainOpen(path); err != nil {
		return false
	}

	return true
}

func Init(path string) error {
	_, err := git.PlainInit(path, false)
	return err
}
