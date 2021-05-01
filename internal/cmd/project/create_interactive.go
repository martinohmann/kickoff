package project

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/ghodss/yaml"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/gitignore"
	"github.com/martinohmann/kickoff/internal/homedir"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/license"
	"github.com/martinohmann/kickoff/internal/prompt"
	"github.com/martinohmann/kickoff/internal/repository"
	"github.com/martinohmann/kickoff/internal/template"
)

func (o *CreateOptions) configureInteractively(config *kickoff.Config) error {
	configureFuncs := []func(*kickoff.Config) error{
		o.configureSkeletons,
		o.configureProject,
		o.configureLicense,
		o.configureGitignoreTemplates,
		o.configureGit,
		o.configureValues,
	}

	for _, configure := range configureFuncs {
		if err := configure(config); err != nil {
			fmt.Fprintln(o.Out)
			return err
		}
	}

	fmt.Fprintln(o.Out)

	return nil
}

func (o *CreateOptions) configureSkeletons(config *kickoff.Config) error {
	if len(o.SkeletonNames) > 0 {
		return nil
	}

	repo, err := o.Repository(o.RepoNames...)
	if err != nil {
		return err
	}

	refs, err := repo.ListSkeletons()
	if err != nil {
		return nil
	}

	options := make([]string, len(refs))
	for i, ref := range refs {
		options[i] = ref.String()
	}

	sort.Strings(options)

	return o.Prompt.AskOne(&survey.MultiSelect{
		Message:  "Select one or more project skeletons",
		Options:  options,
		PageSize: 20,
		VimMode:  true,
	}, &o.SkeletonNames, survey.WithValidator(survey.Required))
}

func (o *CreateOptions) configureProject(config *kickoff.Config) error {
	required := survey.WithValidator(survey.Required)

	if o.ProjectName == "" {
		err := o.Prompt.AskOne(&survey.Input{
			Message: "Project name",
			Help: cmdutil.LongDesc(`
                Project name

                Every project needs a name and so does yours. The name is
                available in templates and can be used to build links related
                to your project, e.g. links to project, CI or docs.`),
		}, &o.ProjectName, required)
		if err != nil {
			return err
		}
	}

	if o.ProjectDir == "" {
		pwd, err := os.Getwd()
		if err != nil {
			return err
		}

		err = o.Prompt.AskOne(&survey.Input{
			Message: "Project directory",
			Default: filepath.Join(pwd, o.ProjectName),
			Suggest: func(toComplete string) []string {
				files, _ := filepath.Glob(homedir.Expand(toComplete) + "*")
				return files
			},
			Help: cmdutil.LongDesc(`
                Project directory

                The directory in which the project will be created.`),
		}, &o.ProjectDir, required)
		if err != nil {
			return err
		}
	}

	if o.ProjectHost == "" {
		err := o.Prompt.AskOne(&survey.Input{
			Message: "Project host",
			Default: config.Project.Host,
			Help: cmdutil.LongDesc(`
                Project host

                To be able to build nice links that are related to the source code repo, e.g. links to
                CI or docs, kickoff needs to know the hostname of your SCM platform.`),
		}, &o.ProjectHost, required)
		if err != nil {
			return err
		}
	}

	if o.ProjectOwner == "" {
		err := o.Prompt.AskOne(&survey.Input{
			Message: "Project owner",
			Default: config.Project.Owner,
			Help: cmdutil.LongDesc(`
                Project owner

                To be able to build nice links that are related to the source code repo, e.g. links to
                CI or docs, kickoff needs to know the username that you use on your SCM platform. 
                The project owner is automatically inserted into license texts if enabled.`),
		}, &o.ProjectOwner, required)
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *CreateOptions) configureLicense(config *kickoff.Config) error {
	if o.License != "" {
		return nil
	}

	client := license.NewClient(o.HTTPClient())

	licenses, err := client.ListLicenses(context.Background())
	if err != nil || len(licenses) == 0 {
		return err
	}

	options := make([]string, 0, len(licenses))
	options = append(options, "None")

	optionMap := make(map[string]string, len(licenses))

	for _, license := range licenses {
		options = append(options, license.Name)
		optionMap[license.Name] = license.Key
	}

	var choice string

	err = o.Prompt.AskOne(&survey.Select{
		Message:  "Choose a license",
		Options:  options,
		PageSize: 20,
		VimMode:  true,
		Help: cmdutil.LongDesc(`
            Open source license

            You can add a open source license to the project. If selected, a
            LICENSE file will be automatically included in the new project
            directory.`),
	}, &choice)
	if err != nil {
		return err
	}

	o.License = optionMap[choice]

	return nil
}

func (o *CreateOptions) configureGitignoreTemplates(config *kickoff.Config) error {
	if len(o.gitignores) > 0 {
		return nil
	}

	client := gitignore.NewClient(o.HTTPClient())

	options, err := client.ListTemplates(context.Background())
	if err != nil || len(options) == 0 {
		return err
	}

	err = o.Prompt.AskOne(&survey.MultiSelect{
		Message:  "Choose gitignore templates",
		Options:  options,
		PageSize: 20,
		VimMode:  true,
		Help: cmdutil.LongDesc(`
            Gitignore templates

            If .gitignore templates are configured, new projects will
            automatically include a .gitignore which is populated with the
            specified templates.`),
	}, &o.gitignores)
	if err != nil {
		return err
	}

	o.Gitignore = strings.Join(o.gitignores, ",")
	return nil
}

func (o *CreateOptions) configureGit(config *kickoff.Config) error {
	if o.InitGit {
		return nil
	}

	return o.Prompt.AskOne(&survey.Confirm{
		Message: "Initialize git in the project directory?",
		Default: o.AutoApprove,
	}, &o.InitGit)
}

func (o *CreateOptions) configureValues(config *kickoff.Config) error {
	if len(o.valuesFiles) > 0 || len(o.rawValues) > 0 {
		return nil
	}

	var edit bool

	err := o.Prompt.AskOne(&survey.Confirm{
		Message: "Edit skeleton values?",
		Default: o.AutoApprove,
		Help: cmdutil.LongDesc(`
            Edit skeleton values

            This allows to edit the values that are passed to template files before rendering them.`),
	}, &edit)
	if err != nil {
		return err
	} else if !edit {
		return nil
	}

	repo, err := o.Repository(o.RepoNames...)
	if err != nil {
		return err
	}

	skeletons, err := repository.LoadSkeletons(repo, o.SkeletonNames)
	if err != nil {
		return err
	}

	merged, err := kickoff.MergeSkeletons(skeletons...)
	if err != nil {
		return err
	}

	if err := o.Values.Merge(merged.Values); err != nil {
		return err
	}

	buf, err := yaml.Marshal(o.Values)
	if err != nil {
		return err
	}

	content := fmt.Sprintf(
		"# Merged values from your kickoff config and the skeletons %s.\n"+
			"# Change them to your needs. To continue, save the file and "+
			"close the editor after you are done.\n%s",
		strings.Join(o.SkeletonNames, ", "), string(buf))

	err = o.Prompt.AskOne(&prompt.Editor{
		Message:         "Open editor",
		FilenamePattern: "*.yaml",
		Default:         content,
		AppendDefault:   true,
		HideDefault:     true,
	}, &content, survey.WithValidator(func(ans interface{}) error {
		var values template.Values
		return yaml.Unmarshal([]byte(ans.(string)), &values)
	}))
	if err != nil {
		return err
	}

	var values template.Values
	if err := yaml.Unmarshal([]byte(content), &values); err != nil {
		return err
	}

	o.Values = values
	return nil
}
