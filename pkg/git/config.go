package git

import (
	"errors"

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
		log.Warn("user.name not found in git config, set it to automatically populate author fields")
	}

	email, err := gitconfig.Global("user.email")
	if err != nil {
		log.Warn("user.email not found in git config, set it to automatically populate email fields")
	}

	githubUser, err := gitconfig.Global("github.user")
	if err != nil {
		log.Warn("github.user not found in git config, set it to automatically populate github user fields")
	}

	return &Config{
		Username:   username,
		Email:      email,
		GitHubUser: githubUser,
	}
}

func EnsureInitialized(path string) error {
	_, err := git.PlainOpen(path)
	if err != nil && errors.Is(err, git.ErrRepositoryNotExists) {
		log.WithField("path", path).Info("initializing git repository")

		_, err = git.PlainInit(path, false)
	}

	return err
}
