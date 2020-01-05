package cmd

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/apex/log"
	"github.com/imdario/mergo"
	"github.com/martinohmann/kickoff/pkg/config"
	"github.com/martinohmann/kickoff/pkg/file"
	"github.com/martinohmann/kickoff/pkg/kickoff"
	"github.com/spf13/cobra"
)

func NewCreateCmd() *cobra.Command {
	o := NewCreateOptions()

	cmd := &cobra.Command{
		Use:   "create <output-dir>",
		Short: "Create project skeletons",
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

	o.AddFlags(cmd)

	return cmd
}

type CreateOptions struct {
	ConfigPath string
	OutputDir  string
	DryRun     bool
	Force      bool

	Config *config.Config
}

func NewCreateOptions() *CreateOptions {
	return &CreateOptions{
		Config: config.NewDefaultConfig(),
	}
}

func (o *CreateOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.DryRun, "dry-run", o.DryRun, "Only print what would be done")
	cmd.Flags().BoolVar(&o.Force, "force", o.Force, "Forces overwrite of existing output directory")
	cmd.Flags().StringVar(&o.ConfigPath, "config", o.ConfigPath, fmt.Sprintf("Path to config file (defaults to %q if the file exists)", config.DefaultConfigPath))

	o.Config.AddFlags(cmd)
}

func (o *CreateOptions) Complete(args []string) (err error) {
	if args[0] != "" {
		o.OutputDir, err = filepath.Abs(args[0])
		if err != nil {
			return err
		}
	}

	if o.ConfigPath == "" && file.Exists(config.DefaultConfigPath) {
		o.ConfigPath = config.DefaultConfigPath
	}

	if o.ConfigPath != "" {
		log.WithField("path", o.ConfigPath).Debugf("loading config file")

		config, err := config.Load(o.ConfigPath)
		if err != nil {
			return err
		}

		err = mergo.Merge(o.Config, config)
		if err != nil {
			return err
		}
	}

	err = o.Config.Complete(o.OutputDir)
	if err != nil {
		return err
	}

	skeletonConfigPath := o.Config.SkeletonConfigPath()

	if file.Exists(skeletonConfigPath) {
		log.WithField("skeleton", o.Config.SkeletonDir()).Debugf("found %s, merging config values", config.SkeletonConfigFile)

		config, err := config.Load(skeletonConfigPath)
		if err != nil {
			return err
		}

		err = mergo.Merge(o.Config, config)
		if err != nil {
			return err
		}
	}

	if o.Config.License == "none" {
		o.Config.License = ""
	}

	return nil
}

func (o *CreateOptions) Validate() error {
	if file.Exists(o.OutputDir) && !o.Force {
		return fmt.Errorf("output-dir %s already exists, add --force to overwrite", o.OutputDir)
	}

	skeletonDir := o.Config.SkeletonDir()

	ok, err := file.IsDirectory(skeletonDir)
	if err != nil {
		return fmt.Errorf("failed to stat skeleton directory: %v", err)
	}

	if !ok {
		return fmt.Errorf("invalid skeleton: %s is not a directory", skeletonDir)
	}

	if o.OutputDir == "" {
		return errors.New("output-dir must not be an empty string")
	}

	return o.Config.Validate()
}

func (o *CreateOptions) Run() error {
	ko := kickoff.New(o.Config, o.DryRun)

	return ko.Create(o.OutputDir)
}
