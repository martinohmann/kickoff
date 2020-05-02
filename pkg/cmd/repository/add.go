package repository

import (
	"errors"
	"fmt"

	"github.com/apex/log"
	"github.com/spf13/cobra"
	"kickoff.run/pkg/cmdutil"
	"kickoff.run/pkg/config"
	"kickoff.run/pkg/skeleton"
)

var (
	ErrEmptyRepositoryName = errors.New("repository name must not be empty")

	ErrEmptyRepositoryURL = errors.New("repository url must not be empty")
)

func NewAddCmd() *cobra.Command {
	o := &AddOptions{}

	cmd := &cobra.Command{
		Use:   "add <name> <url>",
		Short: "Add a skeleton repository to the config",
		Long: cmdutil.LongDesc(`
			Adds a skeleton repository to the config. If a config for the same repository name already exists it will be overridden.`),
		Example: cmdutil.Examples(`
			# Add a new skeleton repository
			kickoff repository add myskeletons /path/to/skeleton/repo`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(args); err != nil {
				return err
			}

			if err := o.Validate(); err != nil {
				return err
			}

			return o.Run()
		},
	}

	cmd.MarkZshCompPositionalArgumentFile(2)

	cmdutil.AddConfigFlag(cmd, &o.ConfigPath)

	return cmd
}

type AddOptions struct {
	cmdutil.ConfigFlags

	RepoName string
	RepoURL  string
}

func (o *AddOptions) Complete(args []string) error {
	o.RepoName = args[0]
	o.RepoURL = args[1]

	return o.ConfigFlags.Complete()
}

func (o *AddOptions) Validate() error {
	if o.RepoName == "" {
		return ErrEmptyRepositoryName
	}

	if o.RepoURL == "" {
		return ErrEmptyRepositoryURL
	}

	return nil
}

func (o *AddOptions) Run() error {
	_, err := skeleton.OpenRepository(o.RepoURL)
	if err != nil {
		return fmt.Errorf("failed to open repository: %v", err)
	}

	o.Repositories[o.RepoName] = o.RepoURL

	err = config.Save(&o.Config, o.ConfigPath)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"name": o.RepoName,
		"url":  o.RepoURL,
	}).Info("repository added")

	return nil
}
