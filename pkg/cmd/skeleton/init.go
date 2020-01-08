package skeleton

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/martinohmann/kickoff/pkg/boilerplate"
	"github.com/martinohmann/kickoff/pkg/file"
	"github.com/martinohmann/kickoff/pkg/skeleton"
	"github.com/spf13/cobra"
)

func NewInitCmd() *cobra.Command {
	o := &InitOptions{}

	cmd := &cobra.Command{
		Use:   "init <output-dir>",
		Short: "Initialize a new skeleton directory",
		Long:  "Initialize a new skeleton directory with some boilerplate to get started",
		Args:  cobra.ExactArgs(1),
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

	o.Out = cmd.OutOrStdout()

	cmd.Flags().BoolVar(&o.Force, "force", o.Force, "Forces overwrite of existing output directory")

	return cmd
}

type InitOptions struct {
	OutputDir string
	Force     bool
	Out       io.Writer
}

func (o *InitOptions) Complete(args []string) (err error) {
	if args[0] != "" {
		o.OutputDir, err = filepath.Abs(args[0])
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *InitOptions) Validate() error {
	if file.Exists(o.OutputDir) && !o.Force {
		return fmt.Errorf("output-dir %s already exists, add --force to overwrite", o.OutputDir)
	}

	if o.OutputDir == "" {
		return errors.New("output-dir must not be an empty string")
	}

	return nil
}

func (o *InitOptions) Run() error {
	log.WithField("path", o.OutputDir).Info("creating skeleton directory")

	err := os.MkdirAll(o.OutputDir, 0755)
	if err != nil {
		return err
	}

	readmeSkelPath := filepath.Join(o.OutputDir, "README.md.skel")

	log.WithField("path", readmeSkelPath).Info("writing README.md.skel")

	err = ioutil.WriteFile(readmeSkelPath, boilerplate.DefaultReadme(), 0644)
	if err != nil {
		return err
	}

	configPath := filepath.Join(o.OutputDir, skeleton.ConfigFile)

	log.WithField("path", configPath).Infof("writing %s", skeleton.ConfigFile)

	err = ioutil.WriteFile(configPath, boilerplate.DefaultConfig(), 0644)
	if err != nil {
		return err
	}

	return nil
}
