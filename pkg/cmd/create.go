package cmd

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/apex/log"
	"github.com/martinohmann/kickoff/pkg/file"
	"github.com/martinohmann/kickoff/pkg/kickoff"
	"github.com/martinohmann/kickoff/pkg/license"
	"github.com/martinohmann/kickoff/pkg/project"
	"github.com/martinohmann/kickoff/pkg/repo"
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

	*kickoff.Config

	rawValues []string
}

func NewCreateOptions() *CreateOptions {
	return &CreateOptions{
		Config: &kickoff.Config{
			Values: template.Values{},
		},
	}
}

func (o *CreateOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.DryRun, "dry-run", o.DryRun, "Only print what would be done")
	cmd.Flags().BoolVar(&o.Force, "force", o.Force, "Forces overwrite of existing output directory")
	cmd.Flags().StringVar(&o.ConfigPath, "config", o.ConfigPath, fmt.Sprintf("Path to config file (defaults to %q if the file exists)", kickoff.DefaultConfigPath))

	cmd.Flags().StringVar(&o.License, "license", o.License, "License to use for the project. If set this will automatically populate the LICENSE file")

	cmd.Flags().StringVar(&o.Project.Name, "project-name", o.Project.Name, "Name of the project. Will be inferred from the output dir if not explicitly set")
	cmd.Flags().StringVar(&o.Project.Author, "project-author", o.Project.Author, "Project author's fullname")
	cmd.Flags().StringVar(&o.Project.Email, "project-email", o.Project.Email, "Project author's e-mail")

	cmd.Flags().StringVar(&o.Git.User, "git-user", o.Git.User, "Git repository user")
	cmd.Flags().StringVar(&o.Git.RepoName, "git-repo-name", o.Git.RepoName, "Git repository name for the project (defaults to the project name)")
	cmd.Flags().StringVar(&o.Git.Host, "git-host", o.Git.Host, "Git repository host")

	cmd.Flags().StringVar(&o.Repo.URL, "repository-url", o.Repo.URL, fmt.Sprintf("URL of the skeleton repository. Can be a local path or remote git repository. (defaults to %q if the directory exists)", repo.DefaultRepositoryURL))

	cmd.Flags().StringArrayVar(&o.rawValues, "set", o.rawValues, "Set custom values of the form key1=value1,key2=value2,deeply.nested.key3=value that are then made available to .skel templates")
}

func (o *CreateOptions) Complete(args []string) (err error) {
	o.Skeleton = args[0]

	if args[1] != "" {
		o.OutputDir, err = filepath.Abs(args[0])
		if err != nil {
			return err
		}
	}

	if o.ConfigPath == "" && file.Exists(kickoff.DefaultConfigPath) {
		o.ConfigPath = kickoff.DefaultConfigPath
	}

	if o.ConfigPath != "" {
		log.WithField("path", o.ConfigPath).Debugf("loading config file")

		err = o.Config.MergeFromFile(o.ConfigPath)
		if err != nil {
			return err
		}
	}

	defaultProjectName := filepath.Base(o.OutputDir)

	o.Project.ApplyDefaults(defaultProjectName)
	o.Git.ApplyDefaults(o.Project.Name)
	o.Repo.ApplyDefaults()

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

	repo, err := repo.Open(o.Repo.URL)
	if err != nil {
		return err
	}

	skeleton, err := repo.Skeleton(o.Skeleton)
	if err != nil {
		return err
	}

	var licenseInfo *license.Info
	if o.License != "none" {
		log.WithField("license", o.License).Debugf("fetching license info from GitHub")

		licenseInfo, err = license.Get(o.License)
		if err == license.ErrLicenseNotFound {
			return fmt.Errorf("license %q not found, use the `licenses` subcommand to get a list of available licenses", o.License)
		} else if err != nil {
			return err
		}
	}

	projectCreator := project.NewCreator(o.Project)

	return projectCreator.Create(skeleton, o.OutputDir, &project.CreateOptions{
		License: licenseInfo,
		DryRun:  o.DryRun,
		Git:     o.Git,
		Values:  o.Values,
	})
}
