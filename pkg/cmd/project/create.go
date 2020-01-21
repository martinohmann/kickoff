package project

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/apex/log"
	"github.com/martinohmann/kickoff/pkg/cmdutil"
	"github.com/martinohmann/kickoff/pkg/file"
	"github.com/martinohmann/kickoff/pkg/project"
	"github.com/martinohmann/kickoff/pkg/skeleton"
	"github.com/spf13/cobra"
	"helm.sh/helm/pkg/strvals"
)

func NewCreateCmd() *cobra.Command {
	o := &CreateOptions{}

	cmd := &cobra.Command{
		Use:   "create <skeleton-name> <output-dir>",
		Short: "Create a project from a skeleton",
		Long: cmdutil.LongDesc(`
			Create a project from a skeleton.`),
		Example: cmdutil.Examples(`
			# Create project
			kickoff project create myskeleton ~/repos/myproject

			# Create project from skeleton in specific repo
			kickoff project create myrepo:myskeleton ~/repos/myproject

			# Create project with license
			kickoff project create myskeleton ~/repos/myproject --license mit

			# Create project with gitignore
			kickoff project create myskeleton ~/repos/myproject --gitignore go,helm,hugo

			# Create project with value overrides
			kickoff project create myskeleton ~/repos/myproject --set travis.enabled=true,mykey=mynewvalue

			# Dry run project creation
			kickoff project create myskeleton ~/repos/myproject --dry-run

			# Forces overwrite of skeleton files in existing project
			kickoff project create myskeleton ~/repos/myproject --force`),
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

	o.AddFlags(cmd)
	o.ConfigFlags.AddFlags(cmd)

	cmdutil.AddForceFlag(cmd, &o.Force)

	return cmd
}

type CreateOptions struct {
	cmdutil.ConfigFlags

	OutputDir string
	Skeleton  string
	DryRun    bool
	Force     bool

	rawValues []string
}

func (o *CreateOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.DryRun, "dry-run", o.DryRun, "Only print what would be done")

	cmd.Flags().StringVar(&o.License, "license", o.License, "License to use for the project. If set this will automatically populate the LICENSE file")
	cmd.Flags().StringVar(&o.Gitignore, "gitignore", o.Gitignore, "Comma-separated list of gitignore template to use for the project. If set this will automatically populate the .gitignore file")

	cmd.Flags().StringVar(&o.Project.Name, "project-name", o.Project.Name, "Name of the project. Will be inferred from the output dir if not explicitly set")
	cmd.Flags().StringVar(&o.Project.Author, "project-author", o.Project.Author, "Project author's fullname")
	cmd.Flags().StringVar(&o.Project.Email, "project-email", o.Project.Email, "Project author's e-mail")

	cmd.Flags().StringVar(&o.Git.User, "git-user", o.Git.User, "Git repository user")
	cmd.Flags().StringVar(&o.Git.RepoName, "git-repo-name", o.Git.RepoName, "Git repository name for the project (defaults to the project name)")
	cmd.Flags().StringVar(&o.Git.Host, "git-host", o.Git.Host, "Git repository host")

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

	defaultProjectName := filepath.Base(o.OutputDir)

	err = o.ConfigFlags.Complete(defaultProjectName)
	if err != nil {
		return err
	}

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
		return fmt.Errorf("output dir %s already exists, add --force to overwrite", o.OutputDir)
	}

	if o.OutputDir == "" {
		return cmdutil.ErrEmptyOutputDir
	}

	if o.Skeleton == "" {
		return cmdutil.ErrEmptySkeletonName
	}

	if o.Git.User == "" {
		return errors.New("--git-user needs to be set as it could not be inferred")
	}

	return nil
}

func (o *CreateOptions) Run() error {
	log.WithField("config", fmt.Sprintf("%#v", o.Config)).Debug("using config")

	repo, err := skeleton.NewMultiRepo(o.Repositories)
	if err != nil {
		return err
	}

	skeleton, err := repo.LoadSkeleton(o.Skeleton)
	if err != nil {
		return err
	}

	return project.Create(skeleton, o.OutputDir, &project.CreateOptions{
		Config: o.Config,
		DryRun: o.DryRun,
	})
}
