package project

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/file"
	"github.com/martinohmann/kickoff/internal/gitignore"
	"github.com/martinohmann/kickoff/internal/license"
	"github.com/martinohmann/kickoff/internal/project"
	"github.com/martinohmann/kickoff/internal/skeleton"
	"github.com/martinohmann/kickoff/internal/template"
	"github.com/spf13/cobra"
	"helm.sh/helm/pkg/strvals"
)

func NewCreateCmd() *cobra.Command {
	o := &CreateOptions{
		TimeoutFlag: cmdutil.NewDefaultTimeoutFlag(),
	}

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

			# Create project with value overrides from files
			kickoff project create myskeleton ~/repos/myproject --values values.yaml --values values2.yaml

			# Create project with value overrides via --set
			kickoff project create myskeleton ~/repos/myproject --set travis.enabled=true,mykey=mynewvalue

			# Dry run project creation
			kickoff project create myskeleton ~/repos/myproject --dry-run

			# Composition of multiple skeletons (comma separated)
			kickoff project create firstskeleton,secondskeleton,thirdskeleton ~/repos/myproject

			# Forces creation of project in existing directory, retaining existing files
			kickoff project create myskeleton ~/repos/myproject --force

			# Forces creation of project in existing directory, overwriting existing files
			kickoff project create myskeleton ~/repos/myproject --force --overwrite`),
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

	o.AddFlags(cmd)
	o.ConfigFlags.AddFlags(cmd)
	o.TimeoutFlag.AddFlag(cmd)

	cmdutil.AddForceFlag(cmd, &o.Force)
	cmdutil.AddOverwriteFlag(cmd, &o.Overwrite)

	return cmd
}

type CreateOptions struct {
	cmdutil.ConfigFlags
	cmdutil.TimeoutFlag

	OutputDir  string
	Skeletons  []string
	DryRun     bool
	Force      bool
	Overwrite  bool
	AllowEmpty bool

	rawValues   []string
	valuesFiles []string
	initGit     bool
}

func (o *CreateOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.DryRun, "dry-run", o.DryRun, "Only print what would be done")

	cmd.Flags().StringVar(&o.Project.Gitignore, "gitignore", o.Project.Gitignore, "Comma-separated list of gitignore template to use for the project. If set this will automatically populate the .gitignore file")
	cmd.Flags().StringVar(&o.Project.Host, "host", o.Project.Host, "Project repository host")
	cmd.Flags().StringVar(&o.Project.License, "license", o.Project.License, "License to use for the project. If set this will automatically populate the LICENSE file")
	cmd.Flags().StringVar(&o.Project.Name, "name", o.Project.Name, "Name of the project. Will be inferred from the output dir if not explicitly set")
	cmd.Flags().StringVar(&o.Project.Owner, "owner", o.Project.Owner, "Project repository owner. This should be the name of the SCM user, e.g. the GitHub user or organization name")

	cmd.Flags().StringArrayVar(&o.valuesFiles, "values", o.valuesFiles, "Load custom values from provided file, making them available to .skel templates. Values passed via --set take precedence")
	cmd.Flags().StringArrayVar(&o.rawValues, "set", o.rawValues, "Set custom values of the form key1=value1,key2=value2,deeply.nested.key3=value that are then made available to .skel templates")

	cmd.Flags().BoolVar(&o.initGit, "init-git", o.initGit, "Initialize git in the project directory")

	cmd.Flags().BoolVar(&o.AllowEmpty, "allow-empty", o.AllowEmpty, "If true, empty files that are the result of template rendering will still be created in the output directory")
}

func (o *CreateOptions) Complete(args []string) (err error) {
	skeletons := args[0]
	outputDir := args[1]

	o.Skeletons = strings.Split(skeletons, ",")

	if outputDir != "" {
		o.OutputDir, err = filepath.Abs(outputDir)
		if err != nil {
			return err
		}
	}

	if o.Project.Name == "" {
		o.Project.Name = filepath.Base(o.OutputDir)
	}

	err = o.ConfigFlags.Complete()
	if err != nil {
		return err
	}

	if len(o.valuesFiles) > 0 {
		for _, path := range o.valuesFiles {
			vals, err := template.LoadValues(path)
			if err != nil {
				return err
			}

			err = o.Values.Merge(vals)
			if err != nil {
				return err
			}
		}
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

	for i, name := range o.Skeletons {
		if name == "" {
			return fmt.Errorf("empty skeleton name index %d", i)
		}
	}

	if o.Project.Owner == "" {
		return errors.New("--owner needs to be set as it could not be inferred")
	}

	return nil
}

func (o *CreateOptions) Run() error {
	log.WithField("config", fmt.Sprintf("%#v", o.Config)).Debug("using config")

	loader, err := skeleton.NewRepositoryAggregateLoader(o.Repositories)
	if err != nil {
		return err
	}

	skeletons, err := loader.LoadSkeletons(o.Skeletons)
	if err != nil {
		return err
	}

	skeleton, err := skeleton.Merge(skeletons...)
	if err != nil {
		return err
	}

	options := &project.CreateOptions{
		DryRun:     o.DryRun,
		Config:     o.Project,
		Values:     o.Values,
		InitGit:    o.initGit,
		Overwrite:  o.Overwrite,
		AllowEmpty: o.AllowEmpty,
	}

	if o.Project.HasLicense() {
		options.License, err = o.fetchLicense(o.Project.License)
		if err != nil {
			return err
		}
	}

	if o.Project.HasGitignore() {
		options.Gitignore, err = o.fetchGitignore(o.Project.Gitignore)
		if err != nil {
			return err
		}
	}

	return project.Create(skeleton, o.OutputDir, options)
}

func (o *CreateOptions) fetchLicense(name string) (*license.Info, error) {
	ctx, cancel := o.TimeoutFlag.Context()
	defer cancel()

	l, err := license.Get(ctx, name)
	if err == license.ErrNotFound {
		return nil, fmt.Errorf("license %q not found, run `kickoff licenses list` to get a list of available licenses", name)
	} else if err != nil {
		return nil, fmt.Errorf("failed to fetch license due to: %v", err)
	}

	return l, nil
}

func (o *CreateOptions) fetchGitignore(template string) (string, error) {
	ctx, cancel := o.TimeoutFlag.Context()
	defer cancel()

	gi, err := gitignore.Get(ctx, template)
	if err == gitignore.ErrNotFound {
		return "", fmt.Errorf("gitignore template %q not found, run `kickoff gitignore list` to get a list of available templates", template)
	} else if err != nil {
		return "", fmt.Errorf("failed to fetch gitignore templates due to: %v", err)
	}

	return gi, nil
}
