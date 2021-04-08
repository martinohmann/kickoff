package skeleton

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/file"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/skeleton"
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
	if o.RepoName == "" {
		return errors.New("repository name must not be empty")
	}

	if o.SkeletonName == "" {
		return errors.New("skeleton name must not be empty")
	}

	if _, ok := o.Repositories[o.RepoName]; !ok {
		return fmt.Errorf("repository with name %q does not exist", o.RepoName)
	}

	return o.ConfigFlags.Validate()
}

// Run creates a new project skeleton in the provided output directory.
func (o *CreateOptions) Run() error {
	repoRef, err := kickoff.ParseRepoRef(o.Repositories[o.RepoName])
	if err != nil {
		return err
	}

	if repoRef.IsRemote() {
		return fmt.Errorf("repository %q is remote. skeletons can only be created in local repositories", o.RepoName)
	}

	skeletonDir := filepath.Join(repoRef.Path, kickoff.SkeletonsDir, o.SkeletonName)

	if file.Exists(skeletonDir) {
		return fmt.Errorf("skeleton %q already exists in repository %q", o.SkeletonName, o.RepoName)
	}

	err = skeleton.Create(skeletonDir)
	if err != nil {
		return err
	}

	fmt.Fprintf(o.Out, "Created new skeleton %q in repository %q\n\n", o.SkeletonName, o.RepoName)
	fmt.Fprintf(o.Out, "You can inspect it by running `kickoff skeleton show %s:%s`.\n", o.RepoName, o.SkeletonName)

	return nil
}
