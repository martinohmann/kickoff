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
	"github.com/martinohmann/kickoff/internal/skeleton"
	"github.com/spf13/cobra"
)

// NewCreateCmd creates a command for creating a local skeleton repository.
func NewCreateCmd(streams cli.IOStreams) *cobra.Command {
	o := &CreateOptions{
		IOStreams:    streams,
		SkeletonName: kickoff.DefaultSkeletonName,
	}

	cmd := &cobra.Command{
		Use:   "create <output-dir>",
		Short: "Create a new skeleton repository",
		Long: cmdutil.LongDesc(`
			Creates a new skeleton repository with a default skeleton to get you started.`),
		Example: cmdutil.Examples(`
			# Create a new repository
			kickoff repository create /repository/output/path

			# Overwrite an existing repository
			kickoff repository create /existing/repository --force`),
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

	cmd.MarkZshCompPositionalArgumentFile(1)

	cmdutil.AddForceFlag(cmd, &o.Force)
	cmd.Flags().StringVar(&o.SkeletonName, "skeleton-name", o.SkeletonName, "Name of the default skeleton that will be created in the new repository.")

	return cmd
}

// CreateOptions holds the options for the create command.
type CreateOptions struct {
	cli.IOStreams
	OutputDir    string
	SkeletonName string
	Force        bool
}

// Complete completes the options for the create command.
func (o *CreateOptions) Complete(args []string) (err error) {
	if args[0] != "" {
		o.OutputDir, err = filepath.Abs(args[0])
		if err != nil {
			return err
		}
	}

	return nil
}

// Validate validates the create options.
func (o *CreateOptions) Validate() error {
	if file.Exists(o.OutputDir) && !o.Force {
		return fmt.Errorf("output dir %s already exists, add --force to overwrite", o.OutputDir)
	}

	ok, err := skeleton.IsInsideSkeletonDir(o.OutputDir)
	if err != nil {
		return err
	}

	if ok {
		return fmt.Errorf("output dir %s is inside a skeleton dir", o.OutputDir)
	}

	if o.OutputDir == "" {
		return cmdutil.ErrEmptyOutputDir
	}

	if o.SkeletonName == "" {
		return errors.New("--skeleton-name must not be empty")
	}

	return nil
}

// Run creates a new skeleton repository in the provided output directory and
// seeds it with a default skeleton.
func (o *CreateOptions) Run() error {
	err := repository.Create(o.OutputDir, o.SkeletonName)
	if err != nil {
		return err
	}

	fmt.Fprintf(o.Out, "Created new skeleton repository in %s\n", o.OutputDir)

	return nil
}
