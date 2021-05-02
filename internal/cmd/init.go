package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/gitignore"
	"github.com/martinohmann/kickoff/internal/homedir"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/license"
	"github.com/martinohmann/kickoff/internal/prompt"
	"github.com/martinohmann/kickoff/internal/repository"
	"github.com/spf13/cobra"
)

var bold = color.New(color.Bold)

// NewInitCmd creates a new command which lets users interactively initialize
// the kickoff configuration.
func NewInitCmd(f *cmdutil.Factory) *cobra.Command {
	o := &InitOptions{
		IOStreams:  f.IOStreams,
		Config:     f.Config,
		ConfigPath: f.ConfigPath,
		HTTPClient: f.HTTPClient,
		Prompt:     f.Prompt,
	}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize kickoff",
		Long: cmdutil.LongDesc(`
			Interactively initialize the kickoff configuration.
		`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.Run()
		},
	}

	return cmd
}

// InitOptions holds the options for the init command.
type InitOptions struct {
	cli.IOStreams

	Config     func() (*kickoff.Config, error)
	HTTPClient func() *http.Client
	Prompt     prompt.Prompt

	ConfigPath string
}

// Run runs the interactive configuration of kickoff.
func (o *InitOptions) Run() error {
	config, err := o.Config()
	if err != nil {
		return err
	}

	configureFuncs := []func(*kickoff.Config) error{
		o.configureProject,
		o.configureLicense,
		o.configureGitignoreTemplates,
		o.configureDefaultSkeletonRepository,
		o.persistConfiguration,
	}

	for _, configure := range configureFuncs {
		err := configure(config)
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *InitOptions) configureProject(config *kickoff.Config) error {
	err := o.Prompt.AskOne(&survey.Input{
		Message: "Default project host",
		Default: config.Project.Host,
		Help: cmdutil.LongDesc(`
            Default project host

            To be able to build nice links that are related to the source code repo, e.g. links to
            CI or docs, kickoff needs to know the hostname of your SCM platform. You can override
            this on project creation.`),
	}, &config.Project.Host)
	if err != nil {
		return err
	}

	return o.Prompt.AskOne(&survey.Input{
		Message: "Default project owner",
		Default: config.Project.Owner,
		Help: cmdutil.LongDesc(`
            Default project owner

            To be able to build nice links that are related to the source code repo, e.g. links to
            CI or docs, kickoff needs to know the username that you use on your SCM platform. You
            can override this on project creation. 
            The project owner is automatically inserted into license texts if enabled.`),
	}, &config.Project.Owner)
}

func (o *InitOptions) configureLicense(config *kickoff.Config) error {
	client := license.NewClient(o.HTTPClient())

	licenses, err := client.ListLicenses(context.Background())
	if err != nil || len(licenses) == 0 {
		return err
	}

	options := make([]string, 0, len(licenses))
	options = append(options, "None")

	optionMap := make(map[string]string, len(licenses))

	var defaultName string

	for _, license := range licenses {
		options = append(options, license.Name)
		optionMap[license.Name] = license.Key
		if license.Key == config.Project.License {
			defaultName = license.Name
		}
	}

	var choice string

	err = o.Prompt.AskOne(&survey.Select{
		Message:  "Default license",
		Options:  options,
		Default:  defaultName,
		PageSize: 20,
		VimMode:  true,
		Help: cmdutil.LongDesc(`
			Open source license

			You can set a default open source license that will be used for all new projects
			if not explicitly overridden on project creation.`),
	}, &choice)
	if err != nil {
		return err
	}

	config.Project.License = optionMap[choice]

	return nil
}

func (o *InitOptions) configureGitignoreTemplates(config *kickoff.Config) error {
	client := gitignore.NewClient(o.HTTPClient())

	options, err := client.ListTemplates(context.Background())
	if err != nil || len(options) == 0 {
		return err
	}

	var choices []string

	err = o.Prompt.AskOne(&survey.MultiSelect{
		Message:  "Default gitignore templates",
		Options:  options,
		Default:  strings.Split(config.Project.Gitignore, ","),
		PageSize: 20,
		VimMode:  true,
		Help: cmdutil.LongDesc(`
			Gitignore templates

			If .gitignore templates are configured, new projects will automatically
			include a .gitignore which is populated with the specified templates.
			Don't worry, you can override this setting on project creation if you want to.`),
	}, &choices)
	if err != nil {
		return err
	}

	config.Project.Gitignore = strings.Join(choices, ",")

	return nil
}

func (o *InitOptions) configureDefaultSkeletonRepository(config *kickoff.Config) error {
	var repoURL string

	err := o.Prompt.AskOne(&survey.Input{
		Message: "Default skeleton repository",
		Default: config.Repositories[kickoff.DefaultRepositoryName],
		Help: cmdutil.LongDesc(`
			Default skeleton repository

			You should at least configure a default skeleton repository to make use of kickoff.
			You can change it or add more repositories at any time if you need to.
		`),
	}, &repoURL)
	if err != nil {
		return err
	}

	ref, err := kickoff.ParseRepoRef(repoURL)
	if err != nil {
		return err
	}

	config.Repositories[kickoff.DefaultRepositoryName] = repoURL

	if ref.IsRemote() {
		return nil
	}

	localPath := ref.LocalPath()

	if _, err := os.Stat(localPath); !os.IsNotExist(err) {
		return err
	}

	var createRepo bool

	err = o.Prompt.AskOne(&survey.Confirm{
		Message: fmt.Sprintf("Skeleton repository %s does not exist, initialize it?", homedir.Collapse(localPath)),
		Default: true,
		Help: cmdutil.LongDesc(`
			Initializing a skeleton repository

			This will create a directory to house your project skeleton and seed it with a "default"
			skeleton which you can customize and use as a starter to create other templates.
		`),
	}, &createRepo)
	if err != nil || !createRepo {
		return err
	}

	repo, err := repository.Create(localPath)
	if err != nil {
		return err
	}

	_, err = repo.CreateSkeleton(kickoff.DefaultSkeletonName)
	return err
}

func (o *InitOptions) persistConfiguration(config *kickoff.Config) error {
	var reviewConfig bool

	err := o.Prompt.AskOne(&survey.Confirm{
		Message: "Do you want to review the configuration before saving it?",
		Default: true,
	}, &reviewConfig)
	if err != nil {
		return err
	}

	if reviewConfig {
		if err := cmdutil.RenderYAML(o.Out, config); err != nil {
			return err
		}
	}

	message := fmt.Sprintf("Save config to %s?", homedir.Collapse(o.ConfigPath))
	if _, err := os.Stat(o.ConfigPath); err == nil {
		message = fmt.Sprintf(
			"There is already a config at %s, do you want to overwrite it?",
			homedir.Collapse(o.ConfigPath),
		)
	}

	var persistConfig bool

	err = o.Prompt.AskOne(&survey.Confirm{Message: message, Default: true}, &persistConfig)
	if err != nil {
		return err
	}

	fmt.Fprintln(o.Out)

	if !persistConfig {
		fmt.Fprintln(o.Out, color.YellowString("!"), "Config was not saved")
		return nil
	}

	err = kickoff.SaveConfig(o.ConfigPath, config)
	if err != nil {
		return err
	}

	fmt.Fprintln(o.Out, color.GreenString("✓"), "Config saved")
	fmt.Fprint(o.Out, "\nHere are some useful commands to get you started:\n\n")
	fmt.Fprintln(o.Out, "- List repositories:", bold.Sprint("kickoff repository list"))
	fmt.Fprintln(o.Out, "- List skeletons:", bold.Sprint("kickoff skeleton list"))
	fmt.Fprintln(o.Out, "- Inspect a skeleton:", bold.Sprint("kickoff skeleton show <skeleton-name>"))
	fmt.Fprintln(o.Out, "- Create new project from skeleton:", bold.Sprint("kickoff project create <name> <skeleton-name>"))

	return nil
}
