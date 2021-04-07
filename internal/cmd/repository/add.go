package repository

import (
	"errors"
	"fmt"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/repository"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// ErrEmptyRepositoryName is returned during validation if the repository
	// name is empty.
	ErrEmptyRepositoryName = errors.New("repository name must not be empty")

	// ErrEmptyRepositoryURL is returned during validation if the repository
	// URL is empty.
	ErrEmptyRepositoryURL = errors.New("repository url must not be empty")
)

// NewAddCmd creates a new command for added a skeleton repository to the
// kickoff config.
func NewAddCmd(streams cli.IOStreams) *cobra.Command {
	o := &AddOptions{
		IOStreams: streams,
	}

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

// AddOptions holds the options for the add command.
type AddOptions struct {
	cli.IOStreams
	cmdutil.ConfigFlags

	RepoName string
	RepoURL  string
}

// Complete completes the add options.
func (o *AddOptions) Complete(args []string) error {
	o.RepoName = args[0]
	o.RepoURL = args[1]

	return o.ConfigFlags.Complete()
}

// Validate validates the options before adding a new repository.
func (o *AddOptions) Validate() error {
	if o.RepoName == "" {
		return ErrEmptyRepositoryName
	}

	if o.RepoURL == "" {
		return ErrEmptyRepositoryURL
	}

	return nil
}

// Run adds a skeleton repository to the kickoff config.
func (o *AddOptions) Run() error {
	_, err := repository.ParseURL(o.RepoURL)
	if err != nil {
		return fmt.Errorf("failed to parse repository URL: %w", err)
	}

	o.Repositories[o.RepoName] = o.RepoURL

	err = kickoff.SaveConfig(o.ConfigPath, &o.Config)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"name": o.RepoName,
		"url":  o.RepoURL,
	}).Info("repository added")

	fmt.Fprintln(o.Out, "Repository added")

	return nil
}
