package repository

import (
	"fmt"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/repository"
	"github.com/spf13/cobra"
)

// NewAddCmd creates a new command for added a skeleton repository to the
// kickoff config.
func NewAddCmd(streams cli.IOStreams) *cobra.Command {
	o := &AddOptions{
		IOStreams:   streams,
		TimeoutFlag: cmdutil.NewDefaultTimeoutFlag(),
	}

	cmd := &cobra.Command{
		Use:   "add <name> <url>",
		Short: "Add a skeleton repository to the config",
		Long: cmdutil.LongDesc(`
			Adds a skeleton repository to the config. If a config for the same repository name already exists it will be overridden.`),
		Example: cmdutil.Examples(`
			# Add a new skeleton repository
			kickoff repository add myskeletons /path/to/skeleton/repo

			# Add a remote skeleton repository in a specific revision
			kickoff repository add myskeletons https://github.com/martinohmann/kickoff-skeletons --revision v1.0.0`),
		Args: cmdutil.ExactNonEmptyArgs(2),
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

	cmd.Flags().StringVar(&o.Revision, "revision", o.Revision, "Revision to checkout. Can be a branch name, tag or commit SHA.")
	cmdutil.AddConfigFlag(cmd, &o.ConfigPath)
	o.TimeoutFlag.AddFlag(cmd)

	return cmd
}

// AddOptions holds the options for the add command.
type AddOptions struct {
	cli.IOStreams
	cmdutil.ConfigFlags
	cmdutil.TimeoutFlag

	RepoName string
	RepoURL  string
	Revision string
}

// Complete completes the add options.
func (o *AddOptions) Complete(args []string) error {
	o.RepoName = args[0]
	o.RepoURL = args[1]

	return o.ConfigFlags.Complete()
}

// Validate validates the options before adding a new repository.
func (o *AddOptions) Validate() error {
	if _, ok := o.Repositories[o.RepoName]; ok {
		return cmdutil.RepositoryAlreadyExistsError(o.RepoName)
	}

	return nil
}

// Run adds a skeleton repository to the kickoff config.
func (o *AddOptions) Run() error {
	ctx, cancel := o.TimeoutFlag.Context()
	defer cancel()

	ref, err := kickoff.ParseRepoRef(o.RepoURL)
	if err != nil {
		return err
	}

	if ref.IsRemote() && o.Revision != "" {
		ref.Revision = o.Revision

		o.RepoURL = ref.String()
	}

	_, err = repository.OpenRef(ctx, *ref, nil)
	if err != nil {
		removeCacheDir(ref)
		return err
	}

	o.Repositories[o.RepoName] = o.RepoURL

	err = kickoff.SaveConfig(o.ConfigPath, &o.Config)
	if err != nil {
		return err
	}

	fmt.Fprintf(o.Out, "Repository added.\n\n")
	fmt.Fprintf(o.Out, "You can inspect it by running `kickoff skeleton list -r %s`.\n", o.RepoName)

	return nil
}
