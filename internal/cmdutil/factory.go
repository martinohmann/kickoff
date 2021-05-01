package cmdutil

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/git"
	"github.com/martinohmann/kickoff/internal/httpcache"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/prompt"
	"github.com/martinohmann/kickoff/internal/repository"
)

// Factory can create instances of commonly needed datastructures like config,
// repository and http client.
type Factory struct {
	IOStreams  cli.IOStreams
	ConfigPath string
	Config     func() (*kickoff.Config, error)
	GitClient  func() git.Client
	HTTPClient func() *http.Client
	Repository func(...string) (kickoff.Repository, error)
	Prompt     prompt.Prompt
}

// NewFactory creates the default *Factory that is passed to commands.
func NewFactory(ioStreams cli.IOStreams) *Factory {
	return NewFactoryWithConfigPath(ioStreams, getConfigPath())
}

// NewFactoryWithConfigPath creates the default *Factory that is passed to commands.
func NewFactoryWithConfigPath(ioStreams cli.IOStreams, configPath string) *Factory {
	var cachedConfig *kickoff.Config

	configFunc := func() (*kickoff.Config, error) {
		if cachedConfig != nil {
			return cachedConfig, nil
		}

		var err error

		cachedConfig, err = kickoff.LoadConfig(configPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) && configPath == kickoff.DefaultConfigPath {
				// A missing default config file is fine. Use the builtin
				// default config.
				cachedConfig = kickoff.DefaultConfig()
				return cachedConfig, nil
			}

			return nil, err
		}

		return cachedConfig, nil
	}

	repositoryFunc := func(names ...string) (kickoff.Repository, error) {
		config, err := configFunc()
		if err != nil {
			return nil, err
		}

		repos := make(map[string]string)

		for _, name := range names {
			if _, ok := config.Repositories[name]; !ok {
				return nil, RepositoryNotConfiguredError(name)
			}

			repos[name] = config.Repositories[name]
		}

		if len(repos) == 0 {
			repos = config.Repositories
		}

		return repository.OpenMap(context.Background(), repos, nil)
	}

	promptFunc := func() prompt.Prompt {
		return prompt.New(func(opts *survey.AskOptions) error {
			if in, ok := ioStreams.In.(terminal.FileReader); ok {
				opts.Stdio.In = in
			} else {
				opts.Stdio.In = os.Stdin
			}

			if out, ok := ioStreams.Out.(terminal.FileWriter); ok {
				opts.Stdio.Out = out
			} else {
				opts.Stdio.Out = os.Stdout
			}

			opts.Stdio.Err = ioStreams.ErrOut
			return nil
		})
	}

	return &Factory{
		ConfigPath: configPath,
		IOStreams:  ioStreams,
		Config:     configFunc,
		GitClient:  git.NewClient,
		HTTPClient: httpcache.NewClient,
		Repository: repositoryFunc,
		Prompt:     promptFunc(),
	}
}

func getConfigPath() string {
	configPath := os.Getenv(kickoff.EnvKeyConfig)
	if configPath != "" {
		return configPath
	}

	return kickoff.DefaultConfigPath
}
