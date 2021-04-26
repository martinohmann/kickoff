package project

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/git"
	"github.com/martinohmann/kickoff/internal/gitignore"
	"github.com/martinohmann/kickoff/internal/homedir"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/license"
	"github.com/martinohmann/kickoff/internal/project"
	"github.com/martinohmann/kickoff/internal/repository"
	"github.com/martinohmann/kickoff/internal/template"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"helm.sh/helm/pkg/strvals"
)

// NewCreateCmd creates a command that can create projects from project
// skeletons using a variety of user-defined options.
func NewCreateCmd(f *cmdutil.Factory) *cobra.Command {
	o := &CreateOptions{
		IOStreams:  f.IOStreams,
		Config:     f.Config,
		GitClient:  f.GitClient,
		HTTPClient: f.HTTPClient,
		Repository: f.Repository,
	}

	cmd := &cobra.Command{
		Use:   "create <name> <skeleton-name> [<skeleton-name>...]",
		Short: "Create a project from one or more skeletons",
		Long: cmdutil.LongDesc(`
			Create a project from one or more skeletons.`),
		Example: cmdutil.Examples(`
			# Create project
			kickoff project create myproject myskeleton

			# Create project from skeleton in specific repo
			kickoff project create myproject myrepo:myskeleton --dir /path/to/project

			# Create project from multiple skeletons (composition)
			kickoff project create myproject repo:myskeleton otherrepo:otherskeleton

			# Create project with license and gitignore templates
			kickoff project create myproject myskeleton --license mit --gitignore go,hugo

			# Create project with value overrides via --set or --values
			kickoff project create myproject myskeleton --set some.val=theval,mykey=mynewvalue --values values.yaml

			# Selectively skip creation of certain files or dirs
			kickoff project create myproject myskeleton --skip-file README.md --skip-file some/dir`),
		Args: cobra.MinimumNArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return cmdutil.SkeletonNames(f, o.RepoNames...), cobra.ShellCompDirectiveDefault
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			o.ProjectName = args[0]
			o.SkeletonNames = args[1:]

			if err := o.Complete(); err != nil {
				return err
			}

			return o.Run()
		},
	}

	o.AddFlags(cmd)

	cmdutil.AddRepositoryFlag(cmd, f, &o.RepoNames)

	cmd.RegisterFlagCompletionFunc("license", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return cmdutil.LicenseNames(f), cobra.ShellCompDirectiveDefault
	})
	cmd.RegisterFlagCompletionFunc("gitignore", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return cmdutil.GitignoreNames(f), cobra.ShellCompDirectiveDefault
	})

	return cmd
}

// CreateOptions holds the options for the create command.
type CreateOptions struct {
	cli.IOStreams

	Config     func() (*kickoff.Config, error)
	GitClient  func() git.Client
	HTTPClient func() *http.Client
	Repository func(...string) (kickoff.Repository, error)

	ProjectName  string
	ProjectDir   string
	ProjectHost  string
	ProjectOwner string
	License      string
	Gitignore    string
	Values       template.Values

	RepoNames      []string
	SkeletonNames  []string
	AutoApprove    bool
	Overwrite      bool
	OverwriteFiles []string
	SkipFiles      []string
	InitGit        bool

	rawValues   []string
	valuesFiles []string
	gitignores  []string
}

// AddFlags adds flags for all project creation options to cmd.
func (o *CreateOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.AutoApprove, "yes", o.AutoApprove, "Auto-approve all prompts")
	cmd.Flags().BoolVar(&o.InitGit, "init-git", o.InitGit, "Initialize git in the project directory")
	cmd.Flags().BoolVar(&o.Overwrite, "overwrite", o.Overwrite, "Overwrite files that are already present in output directory")

	cmd.Flags().StringArrayVar(&o.OverwriteFiles, "overwrite-file", o.OverwriteFiles,
		"Overwrite a specific file in the output directory, if present. File path must be relative to the output directory. "+
			"If file is a dir, present files contained in it will be overwritten")
	cmd.Flags().StringArrayVar(&o.SkipFiles, "skip-file", o.SkipFiles,
		"Skip writing a specific file to the output directory. File path must be relative to the output directory. "+
			"If file is a dir, files contained in it will be skipped as well")
	cmd.Flags().StringArrayVar(&o.rawValues, "set", o.rawValues,
		"Set custom values of the form key1=value1,key2=value2,deeply.nested.key3=value that are then made available to .skel templates")
	cmd.Flags().StringArrayVar(&o.valuesFiles, "values", o.valuesFiles,
		"Load custom values from provided file, making them available to .skel templates. Values passed via --set take precedence")
	cmd.Flags().StringArrayVar(&o.gitignores, "gitignore", o.gitignores,
		"Name of a gitignore template. If provided this will automatically populate the .gitignore file. Can be specified multiple times")
	cmd.Flags().StringVar(&o.License, "license", o.License, "License to use for the project. If set this will automatically populate the LICENSE file")

	cmd.Flags().StringVarP(&o.ProjectDir, "dir", "d", o.ProjectDir, "Custom project directory. If empty the project is created in $PWD/<project-name>")
	cmd.Flags().StringVar(&o.ProjectHost, "host", o.ProjectHost, "Project repository host")
	cmd.Flags().StringVar(&o.ProjectOwner, "owner", o.ProjectOwner, "Project repository owner. This should be the name of the SCM user, e.g. the GitHub user or organization name")
}

// Complete completes the project creation options.
func (o *CreateOptions) Complete() (err error) {
	config, err := o.Config()
	if err != nil {
		return err
	}

	if o.ProjectName == "" {
		return errors.New("project name must not be empty")
	}

	if o.ProjectDir == "" {
		pwd, err := os.Getwd()
		if err != nil {
			return err
		}

		o.ProjectDir = filepath.Join(pwd, o.ProjectName)
	}

	o.ProjectDir, err = filepath.Abs(o.ProjectDir)
	if err != nil {
		return err
	}

	if fi, err := os.Stat(o.ProjectDir); err == nil {
		if !fi.Mode().IsDir() {
			return fmt.Errorf("%s exists but is not a directory", o.ProjectDir)
		}
	}

	if o.ProjectHost == "" {
		o.ProjectHost = config.Project.Host
	}

	if o.ProjectOwner == "" {
		o.ProjectOwner = config.Project.Owner
	}

	if o.ProjectOwner == "" {
		return errors.New("--owner needs to be set as it could not be inferred")
	}

	if o.License == "" {
		o.License = config.Project.License
	}

	o.Gitignore = strings.Join(o.gitignores, ",")
	if o.Gitignore == "" {
		o.Gitignore = config.Project.Gitignore
	}

	o.Values = config.Values

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

// Run loads all project skeletons that the user provided and creates the
// project at the output directory.
func (o *CreateOptions) Run() error {
	repo, err := o.Repository(o.RepoNames...)
	if err != nil {
		return err
	}

	skeletons, err := repository.LoadSkeletons(repo, o.SkeletonNames)
	if err != nil {
		return err
	}

	skeleton, err := kickoff.MergeSkeletons(skeletons...)
	if err != nil {
		return err
	}

	return o.createProject(context.Background(), skeleton)
}

func (o *CreateOptions) createProject(ctx context.Context, skeleton *kickoff.Skeleton) error {
	config := &project.Config{
		Name:           o.ProjectName,
		Host:           o.ProjectHost,
		Owner:          o.ProjectOwner,
		ProjectDir:     o.ProjectDir,
		Overwrite:      o.Overwrite,
		OverwriteFiles: o.OverwriteFiles,
		SkipFiles:      o.SkipFiles,
		Skeleton:       skeleton,
		Values:         o.Values,
	}

	if o.License != "" && o.License != kickoff.NoLicense {
		client := license.NewClient(o.HTTPClient())

		license, err := client.GetLicense(ctx, o.License)
		if err != nil {
			return err
		}

		config.License = license
	}

	if o.Gitignore != "" && o.Gitignore != kickoff.NoGitignore {
		client := gitignore.NewClient(o.HTTPClient())

		template, err := client.GetTemplate(ctx, o.Gitignore)
		if err != nil {
			return err
		}

		config.Gitignore = template
	}

	if err := o.printConfig(config); err != nil {
		return err
	}

	plan, err := project.MakePlan(config)
	if err != nil {
		return err
	}

	o.printPlan(plan)

	if plan.SkipsExisting() {
		fmt.Fprintf(o.Out, "%s Some files will be skipped because they already exist, "+
			"pass %s or %s to overwrite\n\n", color.YellowString("!"), bold.Sprint("--overwrite"), bold.Sprint("--overwrite-file"))
	}

	if plan.IsNoOp() {
		fmt.Fprintf(o.Out, "%s No files to write to %s\n",
			color.YellowString("!"), bold.Sprint(homedir.Collapse(o.ProjectDir)))
		return nil
	}

	if !o.AutoApprove {
		if apply, err := o.confirmApply(); err != nil {
			return err
		} else if !apply {
			return nil
		}
	}

	if err := plan.Apply(); err != nil {
		return err
	}

	o.printSummary(plan)

	if !o.InitGit {
		return nil
	}

	return o.initGitRepository(o.ProjectDir)
}

func (o *CreateOptions) confirmApply() (apply bool, err error) {
	if _, err = os.Stat(o.ProjectDir); err == nil {
		err = survey.AskOne(&survey.Confirm{
			Message: fmt.Sprintf("Project directory %s already exists, still create project?", homedir.Collapse(o.ProjectDir)),
			Default: false,
		}, &apply)
	} else {
		err = survey.AskOne(&survey.Confirm{
			Message: fmt.Sprintf("Create project in %s?", homedir.Collapse(o.ProjectDir)),
			Default: true,
		}, &apply)
	}

	fmt.Fprintln(o.Out)
	return apply, err
}

func (o *CreateOptions) initGitRepository(path string) error {
	log.WithField("path", path).Debug("initializing git repository")

	client := o.GitClient()

	_, err := client.Init(path)
	if errors.Is(err, git.ErrRepositoryAlreadyExists) {
		return nil
	}

	return err
}
