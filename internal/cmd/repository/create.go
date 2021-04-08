package repository

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/file"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/repository"
	"github.com/spf13/cobra"
)

// NewCreateCmd creates a command for creating a local skeleton repository.
func NewCreateCmd(streams cli.IOStreams) *cobra.Command {
	o := &CreateOptions{
		IOStreams:    streams,
		SkeletonName: kickoff.DefaultSkeletonName,
	}

	cmd := &cobra.Command{
		Use:   "create <name> <dir>",
		Short: "Create a new skeleton repository",
		Long: cmdutil.LongDesc(`
			Creates a new skeleton repository with a default skeleton to get you started.`),
		Example: cmdutil.Examples(`
			# Create a new repository
			kickoff repository create myrepo /repository/output/path

            # Create a new repository
			kickoff repository create myrepo /repository/output/path`),
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

	cmd.Flags().StringVarP(&o.SkeletonName, "skeleton-name", "s", o.SkeletonName, "Name of the default skeleton that will be created in the new repository.")
	cmdutil.AddConfigFlag(cmd, &o.ConfigPath)

	return cmd
}

// CreateOptions holds the options for the create command.
type CreateOptions struct {
	cli.IOStreams
	cmdutil.ConfigFlags
	RepoName     string
	OutputDir    string
	SkeletonName string
}

// Complete completes the options for the create command.
func (o *CreateOptions) Complete(args []string) (err error) {
	o.RepoName = args[0]

	o.OutputDir, err = filepath.Abs(args[1])
	if err != nil {
		return err
	}

	return o.ConfigFlags.Complete()
}

// Validate validates the create options.
func (o *CreateOptions) Validate() error {
	if o.RepoName == "" {
		return errors.New("repository name must not be empty")
	}

	if o.SkeletonName == "" {
		return errors.New("--skeleton-name must not be empty")
	}

	if _, ok := o.Repositories[o.RepoName]; ok {
		return fmt.Errorf("repository with name %q already exists", o.RepoName)
	}

	if file.Exists(o.OutputDir) {
		return fmt.Errorf("directory %s already exists", o.OutputDir)
	}

	return o.ConfigFlags.Validate()
}

// Run creates a new skeleton repository in the provided output directory and
// seeds it with a default skeleton.
func (o *CreateOptions) Run() error {
	err := repository.Create(o.OutputDir, o.SkeletonName)
	if err != nil {
		return err
	}

	o.Repositories[o.RepoName] = o.OutputDir

	err = kickoff.SaveConfig(o.ConfigPath, &o.Config)
	if err != nil {
		return err
	}

	fmt.Fprintf(o.Out, "Created new skeleton repository in %s.\n\n", o.OutputDir)
	fmt.Fprintf(o.Out, "You can inspect it by running `kickoff skeleton list -r %s`.\n", o.RepoName)

	return nil
}
