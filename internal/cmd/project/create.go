package project

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/file"
	"github.com/martinohmann/kickoff/internal/git"
	"github.com/martinohmann/kickoff/internal/gitignore"
	"github.com/martinohmann/kickoff/internal/homedir"
	"github.com/martinohmann/kickoff/internal/httpcache"
	"github.com/martinohmann/kickoff/internal/license"
	"github.com/martinohmann/kickoff/internal/project"
	"github.com/martinohmann/kickoff/internal/repository"
	"github.com/martinohmann/kickoff/internal/skeleton"
	"github.com/martinohmann/kickoff/internal/template"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"helm.sh/helm/pkg/strvals"
)

// NewCreateCmd creates a command that can create projects from project
// skeletons using a variety of user-defined options.
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
			kickoff project create myskeleton ~/repos/myproject --force --overwrite

			# Forces creation of project in existing directory, selectively overwriting existing files
			kickoff project create myskeleton ~/repos/myproject --force --overwrite-file README.md`),
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

// CreateOptions holds the options for the create command.
type CreateOptions struct {
	cmdutil.ConfigFlags
	cmdutil.TimeoutFlag

	GitClient       git.Client
	GitignoreClient *gitignore.Client
	LicenseClient   *license.Client

	OutputDir      string
	Skeletons      []string
	DryRun         bool
	Force          bool
	Overwrite      bool
	OverwriteFiles []string
	SkipFiles      []string

	rawValues   []string
	valuesFiles []string
	initGit     bool
}

// AddFlags adds flags for all project creation options to cmd.
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

	cmd.Flags().StringArrayVar(&o.OverwriteFiles, "overwrite-file", o.OverwriteFiles, "Overwrite a specific file in the output directory, if present. File path must be relative to the output directory")
	cmd.Flags().StringArrayVar(&o.SkipFiles, "skip-file", o.SkipFiles, "Skip writing a specific file to the output directory. File path must be relative to the output directory. If file is a dir, files contained in it will be skipped as well")
}

// Complete completes the project creation options.
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

	httpClient := httpcache.NewClient()

	o.GitignoreClient = gitignore.NewClient(httpClient)
	o.LicenseClient = license.NewClient(httpClient)
	o.GitClient = git.NewClient()

	return nil
}

// Validate validates the project creation options.
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

// Run loads all project skeletons that the user provided and creates the
// project at the output directory.
func (o *CreateOptions) Run() error {
	log.WithField("config", fmt.Sprintf("%#v", o.Config)).Debug("using config")

	ctx, cancel := o.TimeoutFlag.Context()
	defer cancel()

	repo, err := repository.NewMultiRepository(o.Repositories)
	if err != nil {
		return err
	}

	skeletons, err := repository.LoadSkeletons(ctx, repo, o.Skeletons)
	if err != nil {
		return err
	}

	skeleton, err := skeleton.Merge(skeletons...)
	if err != nil {
		return err
	}

	builder := project.NewBuilder(o.Project).
		OverwriteAll(o.Overwrite).
		OverwriteFiles(o.OverwriteFiles).
		SkipFiles(o.SkipFiles).
		AddValues(skeleton.Values).
		AddValues(o.Values)

	if o.DryRun {
		builder.WithFilesystem(afero.NewMemMapFs())
	}

	for _, file := range skeleton.Files {
		builder.AddFile(file)
	}

	if o.Project.HasLicense() {
		license, err := o.fetchLicense(ctx, o.Project.License)
		if err != nil {
			return err
		}

		builder.WithLicense(license)
	}

	if o.Project.HasGitignore() {
		gitignore, err := o.fetchGitignore(ctx, o.Project.Gitignore)
		if err != nil {
			return err
		}

		builder.WithGitignore(gitignore)
	}

	outputDir, err := homedir.Collapse(o.OutputDir)
	if err != nil {
		return err
	}

	log.Infof("creating project in %s", outputDir)

	stats, err := builder.Build(o.OutputDir)
	if err != nil {
		return err
	}

	if o.initGit {
		err = o.initGitRepository(o.OutputDir)
		if err != nil {
			return err
		}
	}

	o.logStats(stats)

	return nil
}

func (o *CreateOptions) logStats(stats project.Stats) {
	log.WithFields(log.Fields{
		"dirs.created":  stats.DirsCreated,
		"files.created": stats.FilesCreated,
		"skipped":       stats.Skipped,
	}).Info("project creation complete")

	if stats.Skipped > 0 {
		log.Warn("some targets were skipped, use --overwrite or --overwrite-file to overwrite existing " +
			"files or --allow-empty to allow the creation of empty templates")
	}

	if o.DryRun {
		log.Info("dry-run: no changes have been made")
	}
}

func (o *CreateOptions) fetchLicense(ctx context.Context, name string) (*license.Info, error) {
	l, err := o.LicenseClient.GetLicense(ctx, name)
	if err == license.ErrNotFound {
		return nil, fmt.Errorf("license %q not found, run `kickoff licenses list` to get a list of available licenses", name)
	} else if err != nil {
		return nil, fmt.Errorf("failed to fetch license due to: %v", err)
	}

	return l, nil
}

func (o *CreateOptions) fetchGitignore(ctx context.Context, template string) (string, error) {
	gi, err := o.GitignoreClient.GetTemplate(ctx, template)
	if err == gitignore.ErrNotFound {
		return "", fmt.Errorf("gitignore template %q not found, run `kickoff gitignore list` to get a list of available templates", template)
	} else if err != nil {
		return "", fmt.Errorf("failed to fetch gitignore templates due to: %v", err)
	}

	return gi, nil
}

func (o *CreateOptions) initGitRepository(path string) error {
	log.Info("initializing git repository")

	if !o.DryRun {
		_, err := o.GitClient.Init(path)
		if err != nil && err != git.ErrRepositoryAlreadyExists {
			return err
		}
	}

	return nil
}
