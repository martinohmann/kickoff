package repository

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/config"
	"github.com/martinohmann/kickoff/internal/file"
	"github.com/martinohmann/kickoff/internal/skeleton"
	"github.com/spf13/cobra"
)

func NewCreateCmd() *cobra.Command {
	o := &CreateOptions{SkeletonName: config.DefaultSkeletonName}

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

type CreateOptions struct {
	OutputDir    string
	SkeletonName string
	Force        bool
}

func (o *CreateOptions) Complete(args []string) (err error) {
	if args[0] != "" {
		o.OutputDir, err = filepath.Abs(args[0])
		if err != nil {
			return err
		}
	}

	return nil
}

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

func (o *CreateOptions) Run() error {
	return skeleton.CreateRepository(o.OutputDir, o.SkeletonName)
}
