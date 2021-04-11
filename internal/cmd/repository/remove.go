package repository

import (
	"fmt"
	"os"
	"strings"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/kickoff"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewRemoveCmd creates a command for removing skeleton repositories from the
// config.
func NewRemoveCmd(streams cli.IOStreams) *cobra.Command {
	o := &RemoveOptions{
		IOStreams: streams,
	}

	cmd := &cobra.Command{
		Use:     "remove <name>",
		Aliases: []string{"rm"},
		Short:   "Remove a skeleton repository from the config",
		Long: cmdutil.LongDesc(`
			Removes a skeleton repository from the config. Does not remove local repositories from disk, but cleans up the local cache dir for remote repositories.`),
		Example: cmdutil.Examples(`
			# Remove a skeleton repository
			kickoff repository remove myrepo`),
		Args: cobra.ExactArgs(1),
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

	cmdutil.AddConfigFlag(cmd, &o.ConfigPath)

	return cmd
}

// RemoveOptions holds the options for the remove command.
type RemoveOptions struct {
	cli.IOStreams
	cmdutil.ConfigFlags

	RepoName string
}

// Complete completes the remove options.
func (o *RemoveOptions) Complete(args []string) error {
	o.RepoName = args[0]

	return o.ConfigFlags.Complete()
}

// Validate validates the remove options.
func (o *RemoveOptions) Validate() error {
	if o.RepoName == "" {
		return ErrEmptyRepositoryName
	}

	_, ok := o.Repositories[o.RepoName]
	if !ok {
		return cmdutil.RepositoryNotConfiguredError{Name: o.RepoName}
	}

	return nil
}

// Run removes a skeleton repository from the config.
func (o *RemoveOptions) Run() error {
	repoRef, err := kickoff.ParseRepoRef(o.Repositories[o.RepoName])
	if err != nil {
		return err
	}

	if repoRef.IsRemote() {
		localPath := repoRef.LocalPath()

		// Prevent removal of anything outside of the local user cache dir.
		// Remote repos should never ever reside outside of the user cache dir.
		// If they do this is a programmer error.
		if !strings.HasPrefix(localPath, kickoff.LocalRepositoryCacheDir) {
			log.WithField("path", localPath).
				Panic("unexpected repository location: found remote repository cache outside of user cache dir, refusing to delete")
		}

		log.WithField("path", localPath).Debug("deleting repository cache dir")

		if err := os.RemoveAll(localPath); err != nil {
			return fmt.Errorf("failed to delete repository cache dir: %w", err)
		}
	}

	delete(o.Repositories, o.RepoName)

	err = kickoff.SaveConfig(o.ConfigPath, &o.Config)
	if err != nil {
		return err
	}

	fmt.Fprintln(o.Out, "Repository removed.")

	return nil
}
