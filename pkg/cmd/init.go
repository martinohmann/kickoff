package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/apex/log"
	"github.com/ghodss/yaml"
	"github.com/martinohmann/kickoff/pkg/cli"
	"github.com/martinohmann/kickoff/pkg/cmdutil"
	"github.com/martinohmann/kickoff/pkg/config"
	"github.com/martinohmann/kickoff/pkg/file"
	"github.com/martinohmann/kickoff/pkg/gitignore"
	"github.com/martinohmann/kickoff/pkg/license"
	"github.com/martinohmann/kickoff/pkg/skeleton"
	"github.com/spf13/cobra"
)

func NewInitCmd(streams cli.IOStreams) *cobra.Command {
	o := &InitOptions{
		IOStreams:   streams,
		TimeoutFlag: cmdutil.NewDefaultTimeoutFlag(),
	}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize kickoff",
		Long: cmdutil.LongDesc(`
			Interactively initialize the kickoff configuration.
		`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Complete()
			if err != nil {
				return err
			}

			return o.Run()
		},
	}

	o.AllowMissingConfig()
	o.TimeoutFlag.AddFlag(cmd)
	cmdutil.AddConfigFlag(cmd, &o.ConfigPath)

	return cmd
}

type InitOptions struct {
	cli.IOStreams
	cmdutil.ConfigFlags
	cmdutil.TimeoutFlag
}

func (o *InitOptions) Complete() (err error) {
	err = o.ConfigFlags.Complete()
	if err != nil {
		return err
	}

	if o.ConfigPath == "" {
		o.ConfigPath = config.DefaultConfigPath
	}

	o.ConfigPath, err = filepath.Abs(o.ConfigPath)

	return err
}

func (o *InitOptions) Run() error {
	configureFuncs := []func() error{
		o.configureProject,
		o.configureLicense,
		o.configureGitignoreTemplates,
		o.configureDefaultSkeletonRepository,
		o.persistConfiguration,
	}

	for _, configure := range configureFuncs {
		err := configure()
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *InitOptions) configureProject() error {
	questions := []*survey.Question{
		{
			Name: "host",
			Prompt: &survey.Input{
				Message: "Default project host",
				Default: o.Project.Host,
				Help: cmdutil.LongDesc(`
					Default project host

					To be able to build nice links that are related to the source code repo, e.g. links to
					CI or docs, kickoff needs to know the hostname of your SCM platform. You can override
					this on project creation.
				`),
			},
		},
		{
			Name: "owner",
			Prompt: &survey.Input{
				Message: "Default project owner",
				Default: o.Project.Owner,
				Help: cmdutil.LongDesc(`
					Default project owner

					To be able to build nice links that are related to the source code repo, e.g. links to
					CI or docs, kickoff needs to know the username that you use on your SCM platform. You
					can override this on project creation. 
					The project owner is automatically inserted into license texts if enabled.
				`),
			},
		},
	}

	return survey.Ask(questions, &o.Project)
}

func (o *InitOptions) configureLicense() error {
	var (
		licensesAvailable bool
		licenseOptions    []string
		licenseMap        map[string]string
	)

	ctx, cancel := o.TimeoutFlag.Context()
	defer cancel()

	licenses, err := license.List(ctx)
	if err != nil {
		log.Debugf("skipping license configuration due to: %v", err)
	} else if len(licenses) > 0 {
		licensesAvailable = true

		licenseMap = make(map[string]string)

		for _, license := range licenses {
			licenseOptions = append(licenseOptions, license.Name)
			licenseMap[license.Name] = license.Key
		}
	}

	if !licensesAvailable {
		return nil
	}

	var chooseLicense bool

	err = survey.AskOne(&survey.Confirm{
		Message: "Do you want to set a default project license?",
		Default: false,
		Help: cmdutil.LongDesc(`
			Open source license

			You can set a default open source license that will be used for all new projects
			if not explicitly overridden on project creation.
		`),
	}, &chooseLicense)
	if err != nil {
		return err
	}

	if !chooseLicense {
		o.Project.License = config.NoLicense
		return nil
	}

	var chosenLicense string

	err = survey.AskOne(&survey.Select{
		Message:  "Choose a license",
		Options:  licenseOptions,
		PageSize: 20,
		VimMode:  true,
		Help: cmdutil.LongDesc(`
			Open source license

			You can set a default open source license that will be used for all new projects
			if not explicitly overridden on project creation.
		`),
	}, &chosenLicense)
	if err != nil {
		return err
	}

	o.Project.License = licenseMap[chosenLicense]

	return nil
}

func (o *InitOptions) configureGitignoreTemplates() error {
	var gitignoresAvailable bool

	ctx, cancel := o.TimeoutFlag.Context()
	defer cancel()

	gitignoreOptions, err := gitignore.List(ctx)
	if err != nil {
		log.Debugf("skipping gitignore configuration due to: %v", err)
	} else if len(gitignoreOptions) > 0 {
		gitignoresAvailable = true
	}

	if !gitignoresAvailable {
		return nil
	}

	var selectGitignores bool

	err = survey.AskOne(&survey.Confirm{
		Message: "Do you want to select default .gitignore templates?",
		Default: false,
		Help: cmdutil.LongDesc(`
			Gitignore templates

			If .gitignore templates are configured, new projects will automatically
			include a .gitignore which is populated with the specified templates.
			Don't worry, you can override this setting on project creation if you want to.
		`),
	}, &selectGitignores)
	if err != nil {
		return err
	}

	if !selectGitignores {
		o.Project.Gitignore = config.NoGitignore
		return nil
	}

	var selectedGitignores []string

	err = survey.AskOne(&survey.MultiSelect{
		Message:  "Choose gitignore templates",
		Options:  gitignoreOptions,
		PageSize: 20,
		VimMode:  true,
		Help: cmdutil.LongDesc(`
			Gitignore templates

			If .gitignore templates are configured, new projects will automatically
			include a .gitignore which is populated with the specified templates.
			Don't worry, you can override this setting on project creation if you want to.
		`),
	}, &selectedGitignores)
	if err != nil {
		return err
	}

	o.Project.Gitignore = strings.Join(selectedGitignores, ",")

	return nil
}

func (o *InitOptions) configureDefaultSkeletonRepository() error {
	var repoURL string

	err := survey.AskOne(&survey.Input{
		Message: "Default skeleton repository",
		Default: o.Repositories[config.DefaultRepositoryName],
		Help: cmdutil.LongDesc(`
			Default skeleton repository

			You should at least configure a default skeleton repository to make use of kickoff.
			You can change it or add more repositories at any time if you need to.
		`),
	}, &repoURL)
	if err != nil {
		return err
	}

	info, err := skeleton.ParseRepositoryURL(repoURL)
	if err != nil {
		return err
	}

	o.Repositories[config.DefaultRepositoryName] = repoURL

	if !info.Local {
		return nil
	}

	localPath := info.LocalPath()

	if file.Exists(localPath) {
		return nil
	}

	var createRepo bool

	err = survey.AskOne(&survey.Confirm{
		Message: fmt.Sprintf("Skeleton repository %s does not exist, initialize it?", localPath),
		Default: true,
		Help: cmdutil.LongDesc(`
			Initializing a skeleton repository

			This will create a directory to house your project skeleton and seed it with a "default"
			skeleton which you can customize and use as a starter to create other templates.
		`),
	}, &createRepo)
	if err != nil {
		return err
	}

	if !createRepo {
		return nil
	}

	return skeleton.CreateRepository(localPath, config.DefaultSkeletonName)
}

func (o *InitOptions) persistConfiguration() error {
	var reviewConfig bool

	err := survey.AskOne(&survey.Confirm{
		Message: "Do you want to review the configuration before saving it?",
		Default: true,
	}, &reviewConfig)
	if err != nil {
		return err
	}

	if reviewConfig {
		buf, err := yaml.Marshal(o.Config)
		if err != nil {
			return err
		}

		fmt.Fprintf(o.Out, "\n---\n%s\n", string(buf))
	}

	message := fmt.Sprintf("Save config to %s?", o.ConfigPath)
	if file.Exists(o.ConfigPath) {
		message = fmt.Sprintf(
			"There is already a config at %s, do you want to overwrite it?",
			o.ConfigPath,
		)
	}

	var persistConfig bool

	err = survey.AskOne(&survey.Confirm{Message: message, Default: true}, &persistConfig)
	if err != nil {
		return err
	}

	if !persistConfig {
		log.Error("did not save config")
		return nil
	}

	log.WithField("path", o.ConfigPath).Info("writing config")

	return config.Save(&o.Config, o.ConfigPath)
}
