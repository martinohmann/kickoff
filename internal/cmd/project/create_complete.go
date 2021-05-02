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
	"github.com/martinohmann/kickoff/internal/repository"
	"github.com/martinohmann/kickoff/internal/template"
	"helm.sh/helm/pkg/strvals"
)

func (o *CreateOptions) complete(config *kickoff.Config) error {
	completeFuncs := []func(*kickoff.Config) error{
		o.completeSkeletonNames,
		o.completeProjectName,
		o.completeProjectDir,
		o.completeProjectHost,
		o.completeProjectOwner,
		o.completeLicense,
		o.completeGitignoreTemplates,
		o.completeGitInit,
		o.completeValues,
	}

	for _, complete := range completeFuncs {
		if err := complete(config); err != nil {
			return err
		}
	}

	return nil
}

func (o *CreateOptions) completeSkeletonNames(config *kickoff.Config) error {
	if len(o.SkeletonNames) > 0 && !o.Interactive {
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
		Default:  o.SkeletonNames,
		PageSize: 20,
		VimMode:  true,
	}, &o.SkeletonNames, survey.WithValidator(survey.Required))
}

func (o *CreateOptions) completeProjectName(config *kickoff.Config) error {
	if o.ProjectName != "" && !o.Interactive {
		return nil
	}

	return o.Prompt.AskOne(&survey.Input{
		Message: "Project name",
		Default: o.ProjectName,
		Help: cmdutil.LongDesc(`
            Project name

            Every project needs a name and so does yours. The name is
            available in templates and can be used to build links related
            to your project, e.g. links to project, CI or docs.`),
	}, &o.ProjectName, survey.WithValidator(survey.Required))
}

func (o *CreateOptions) completeProjectDir(config *kickoff.Config) (err error) {
	if o.ProjectDir == "" {
		pwd, err := os.Getwd()
		if err != nil {
			return err
		}

		o.ProjectDir = filepath.Join(pwd, o.ProjectName)
	}

	if o.Interactive {
		err := o.Prompt.AskOne(&survey.Input{
			Message: "Project directory",
			Default: o.ProjectDir,
			Suggest: func(toComplete string) []string {
				files, _ := filepath.Glob(homedir.Expand(toComplete) + "*")
				return files
			},
			Help: cmdutil.LongDesc(`
                Project directory

                The directory in which the project will be created.`),
		}, &o.ProjectDir, survey.WithValidator(survey.Required), survey.WithValidator(func(ans interface{}) error {
			return isDirOrNonexistent(ans.(string))
		}))
		if err != nil {
			return err
		}
	}

	o.ProjectDir, err = filepath.Abs(o.ProjectDir)
	if err != nil {
		return err
	}

	return isDirOrNonexistent(o.ProjectDir)
}

func isDirOrNonexistent(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	} else if !fi.Mode().IsDir() {
		return fmt.Errorf("%s exists but is not a directory", path)
	}

	return nil
}

func (o *CreateOptions) completeProjectHost(config *kickoff.Config) error {
	if o.ProjectHost == "" {
		o.ProjectHost = config.Project.Host
	}

	if o.Interactive || o.ProjectHost == "" {
		return o.Prompt.AskOne(&survey.Input{
			Message: "Project host",
			Default: o.ProjectHost,
			Help: cmdutil.LongDesc(`
                Project host

                To be able to build nice links that are related to the source code repo, e.g. links to
                CI or docs, kickoff needs to know the hostname of your SCM platform.`),
		}, &o.ProjectHost, survey.WithValidator(survey.Required))
	}

	return nil
}

func (o *CreateOptions) completeProjectOwner(config *kickoff.Config) error {
	if o.ProjectOwner == "" {
		o.ProjectOwner = config.Project.Owner
	}

	if o.Interactive || o.ProjectOwner == "" {
		return o.Prompt.AskOne(&survey.Input{
			Message: "Project owner",
			Default: o.ProjectOwner,
			Help: cmdutil.LongDesc(`
                Project owner

                To be able to build nice links that are related to the source code repo, e.g. links to
                CI or docs, kickoff needs to know the username that you use on your SCM platform. 
                The project owner is automatically inserted into license texts if enabled.`),
		}, &o.ProjectOwner, survey.WithValidator(survey.Required))
	}

	return nil
}

func (o *CreateOptions) completeLicense(config *kickoff.Config) error {
	if o.License == "" {
		o.License = config.Project.License
	}

	if !o.Interactive {
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

	var defaultName string

	for _, license := range licenses {
		options = append(options, license.Name)
		optionMap[license.Name] = license.Key
		if o.License == license.Key {
			defaultName = license.Name
		}
	}

	var choice string

	err = o.Prompt.AskOne(&survey.Select{
		Message:  "Choose a license",
		Options:  options,
		Default:  defaultName,
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

func (o *CreateOptions) completeGitignoreTemplates(config *kickoff.Config) error {
	if len(o.gitignores) == 0 {
		o.gitignores = strings.Split(config.Project.Gitignore, ",")
	}

	if o.Interactive {
		client := gitignore.NewClient(o.HTTPClient())

		options, err := client.ListTemplates(context.Background())
		if err != nil || len(options) == 0 {
			return err
		}

		var choices []string

		err = o.Prompt.AskOne(&survey.MultiSelect{
			Message:  "Choose gitignore templates",
			Options:  options,
			Default:  o.gitignores,
			PageSize: 20,
			VimMode:  true,
			Help: cmdutil.LongDesc(`
                Gitignore templates

                If .gitignore templates are configured, new projects will
                automatically include a .gitignore which is populated with the
                specified templates.`),
		}, &choices)
		if err != nil {
			return err
		}

		o.gitignores = choices
	}

	o.Gitignore = strings.Join(o.gitignores, ",")
	return nil
}

func (o *CreateOptions) completeGitInit(config *kickoff.Config) error {
	if o.InitGit && !o.Interactive {
		return nil
	}

	return o.Prompt.AskOne(&survey.Confirm{
		Message: "Initialize git in the project directory?",
		Default: o.InitGit,
	}, &o.InitGit)
}

func (o *CreateOptions) completeValues(config *kickoff.Config) error {
	o.Values = config.Values

	for _, path := range o.valuesFiles {
		vals, err := template.LoadValues(path)
		if err != nil {
			return err
		}

		if err := o.Values.Merge(vals); err != nil {
			return err
		}
	}

	for _, rawValues := range o.rawValues {
		if err := strvals.ParseInto(rawValues, o.Values); err != nil {
			return err
		}
	}

	if !o.Interactive {
		return nil
	}

	var edit bool

	err := o.Prompt.AskOne(&survey.Confirm{
		Message: "Edit skeleton values?",
		Default: true,
		Help: cmdutil.LongDesc(`
            Edit skeleton values

            This allows to edit the values that are passed to template files before rendering them.`),
	}, &edit)
	if err != nil || !edit {
		return err
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

	o.Values, err = template.MergeValues(merged.Values, o.Values)
	if err != nil {
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

	err = o.Prompt.AskOne(&survey.Editor{
		Message:       "Open editor",
		FileName:      "*.yaml",
		Default:       content,
		AppendDefault: true,
		HideDefault:   true,
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
