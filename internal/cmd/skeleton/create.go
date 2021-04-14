package skeleton

import (
	"fmt"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/repository"
	"github.com/spf13/cobra"
)

// NewCreateCmd creates a command for creating project skeletons.
func NewCreateCmd(streams cli.IOStreams) *cobra.Command {
	o := &CreateOptions{
		IOStreams: streams,
	}

	cmd := &cobra.Command{
		Use:   "create <repo-name> <skeleton-name>",
		Short: "Create a new skeleton in a local repository",
		Long: cmdutil.LongDesc(`
			Creates a new skeleton directory in a local repository with some boilerplate to get started.`),
		Example: cmdutil.Examples(`
			# Create a new skeleton in myrepo
			kickoff skeleton create myrepo myskeleton`),
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

	cmdutil.AddConfigFlag(cmd, &o.ConfigPath)

	return cmd
}

// CreateOptions holds the options for the create command.
type CreateOptions struct {
	cli.IOStreams
	cmdutil.ConfigFlags

	RepoName     string
	SkeletonName string
}

// Complete completes the create options.
func (o *CreateOptions) Complete(args []string) (err error) {
	o.RepoName = args[0]
	o.SkeletonName = args[1]

	return o.ConfigFlags.Complete()
}

// Validate validates the create options.
func (o *CreateOptions) Validate() error {
	if _, ok := o.Repositories[o.RepoName]; !ok {
		return cmdutil.RepositoryNotConfiguredError(o.RepoName)
	}

	return nil
}

// Run creates a new project skeleton in the provided output directory.
func (o *CreateOptions) Run() error {
	repoRef, err := kickoff.ParseRepoRef(o.Repositories[o.RepoName])
	if err != nil {
		return err
	}

	repoRef.Name = o.RepoName

	err = repository.CreateSkeleton(repoRef, o.SkeletonName)
	if err != nil {
		return err
	}

	fmt.Fprintf(o.Out, "Created new skeleton %q in repository %q\n\n", o.SkeletonName, o.RepoName)
	fmt.Fprintf(o.Out, "You can inspect it by running `kickoff skeleton show %s:%s`.\n", o.RepoName, o.SkeletonName)

	return nil
}
