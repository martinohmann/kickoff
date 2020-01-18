package skeleton

import (
	"fmt"
	"path/filepath"

	"github.com/martinohmann/kickoff/pkg/cmdutil"
	"github.com/martinohmann/kickoff/pkg/file"
	"github.com/martinohmann/kickoff/pkg/skeleton"
	"github.com/spf13/cobra"
)

func NewCreateCmd() *cobra.Command {
	o := &CreateOptions{}

	cmd := &cobra.Command{
		Use:   "create <output-dir>",
		Short: "Create a new skeleton directory",
		Long: cmdutil.LongDesc(`
			Creates a new skeleton directory with some boilerplate to get started.`),
		Example: cmdutil.Examples(`
			# Create a new skeleton
			kickoff skeleton create /skeleton/output/path

			# Overwrite an existing skeleton
			kickoff skeleton create /existing/skeleton --force`),
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

	cmdutil.AddForceFlag(cmd, &o.Force)

	return cmd
}

type CreateOptions struct {
	OutputDir string
	Force     bool
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

	return nil
}

func (o *CreateOptions) Run() error {
	return skeleton.Create(o.OutputDir)
}
