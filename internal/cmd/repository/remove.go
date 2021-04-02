package repository

import (
	"fmt"

	"github.com/apex/log"
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/config"
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
			Removes a skeleton repository from the config.`),
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
		return fmt.Errorf("repository %q not configured", o.RepoName)
	}

	return nil
}

// Run removes a skeleton repository from the config.
func (o *RemoveOptions) Run() error {
	delete(o.Repositories, o.RepoName)

	err := config.Save(&o.Config, o.ConfigPath)
	if err != nil {
		return err
	}

	log.WithField("name", o.RepoName).Info("repository removed")

	fmt.Fprintln(o.Out, "Repository removed")

	return nil
}
