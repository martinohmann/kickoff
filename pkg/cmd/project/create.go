package project

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/apex/log"
	"github.com/martinohmann/kickoff/pkg/config"
	"github.com/martinohmann/kickoff/pkg/file"
	"github.com/martinohmann/kickoff/pkg/project"
	"github.com/martinohmann/kickoff/pkg/skeleton"
	"github.com/martinohmann/kickoff/pkg/template"
	"github.com/spf13/cobra"
	"helm.sh/helm/pkg/strvals"
)

func NewCreateCmd() *cobra.Command {
	o := NewCreateOptions()

	cmd := &cobra.Command{
		Use:   "create <skeleton-name> <output-dir>",
		Short: "Create a project from a skeleton",
		Args:  cobra.ExactArgs(2),
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
	Skeleton   string
	DryRun     bool
	Force      bool

	config.Config

	rawValues []string
}

func NewCreateOptions() *CreateOptions {
	return &CreateOptions{
		Config: config.Config{
			Values: template.Values{},
		},
	}
}

func (o *CreateOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.DryRun, "dry-run", o.DryRun, "Only print what would be done")
	cmd.Flags().BoolVar(&o.Force, "force", o.Force, "Forces overwrite of existing output directory")
	cmd.Flags().StringVar(&o.ConfigPath, "config", o.ConfigPath, fmt.Sprintf("Path to config file (defaults to %q if the file exists)", config.DefaultConfigPath))

	cmd.Flags().StringVar(&o.License, "license", o.License, "License to use for the project. If set this will automatically populate the LICENSE file")

	cmd.Flags().StringVar(&o.Project.Name, "project-name", o.Project.Name, "Name of the project. Will be inferred from the output dir if not explicitly set")
	cmd.Flags().StringVar(&o.Project.Author, "project-author", o.Project.Author, "Project author's fullname")
	cmd.Flags().StringVar(&o.Project.Email, "project-email", o.Project.Email, "Project author's e-mail")

	cmd.Flags().StringVar(&o.Git.User, "git-user", o.Git.User, "Git repository user")
	cmd.Flags().StringVar(&o.Git.RepoName, "git-repo-name", o.Git.RepoName, "Git repository name for the project (defaults to the project name)")
	cmd.Flags().StringVar(&o.Git.Host, "git-host", o.Git.Host, "Git repository host")

	cmd.Flags().StringVar(&o.Skeletons.RepositoryURL, "repository-url", o.Skeletons.RepositoryURL, fmt.Sprintf("URL of the skeleton repository. Can be a local path or remote git repository. (defaults to %q if the directory exists)", config.DefaultSkeletonRepositoryURL))

	cmd.Flags().StringArrayVar(&o.rawValues, "set", o.rawValues, "Set custom values of the form key1=value1,key2=value2,deeply.nested.key3=value that are then made available to .skel templates")
}

func (o *CreateOptions) Complete(args []string) (err error) {
	o.Skeleton = args[0]

	if args[1] != "" {
		o.OutputDir, err = filepath.Abs(args[1])
		if err != nil {
			return err
		}
	}

	if o.ConfigPath == "" && file.Exists(config.DefaultConfigPath) {
		o.ConfigPath = config.DefaultConfigPath
	}

	if o.ConfigPath != "" {
		log.WithField("path", o.ConfigPath).Debugf("loading config file")

		err = o.Config.MergeFromFile(o.ConfigPath)
		if err != nil {
			return err
		}
	}

	defaultProjectName := filepath.Base(o.OutputDir)

	o.ApplyDefaults(defaultProjectName)

	if len(o.rawValues) > 0 {
		for _, rawValues := range o.rawValues {
			err = strvals.ParseInto(rawValues, o.Values)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (o *CreateOptions) Validate() error {
	if file.Exists(o.OutputDir) && !o.Force {
		return fmt.Errorf("output-dir %s already exists, add --force to overwrite", o.OutputDir)
	}

	if o.OutputDir == "" {
		return errors.New("output-dir must not be an empty string")
	}

	if o.Skeleton == "" {
		return fmt.Errorf("skeleton name must be provided")
	}

	if o.Git.User == "" {
		return fmt.Errorf("--git-user needs to be set as it could not be inferred")
	}

	return nil
}

func (o *CreateOptions) Run() error {
	log.WithField("config", fmt.Sprintf("%#v", o.Config)).Debug("using config")

	repo, err := skeleton.OpenRepository(o.Skeletons.RepositoryURL)
	if err != nil {
		return err
	}

	skeleton, err := repo.Skeleton(o.Skeleton)
	if err != nil {
		return err
	}

	return project.Create(skeleton, o.OutputDir, &project.CreateOptions{
		Config: o.Config,
		DryRun: o.DryRun,
	})
}
