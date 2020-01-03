package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/apex/log"
	"github.com/martinohmann/skeleton-go/pkg/file"
	"github.com/martinohmann/skeleton-go/pkg/git"
	"github.com/martinohmann/skeleton-go/pkg/license"
	"github.com/spf13/cobra"
)

const skeletonBase = "_skeleton"

func newCreateCommand() *cobra.Command {
	o, err := NewDefaultCreateOptions()
	if err != nil {
		panic(err)
	}

	cmd := &cobra.Command{
		Use:   "create <output-dir>",
		Short: "Create golang project skeletons",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(cmd, args); err != nil {
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
	Author      string
	Email       string
	ProjectName string
	License     string

	GitHubUser    string
	GitHubRepo    string
	GitHubRepoURL string

	SkeletonPath string
	OutputDir    string

	DryRun bool
	Force  bool

	LicenseInfo *license.Info
}

func NewDefaultCreateOptions() (*CreateOptions, error) {
	var skeletonPath string
	_, err := os.Stat(skeletonBase)
	if err == nil {
		skeletonPath, err = filepath.Abs(skeletonBase)
		if err != nil {
			return nil, err
		}
	}

	gitConfig := git.GlobalConfig()

	o := &CreateOptions{
		GitHubUser:   gitConfig.GitHubUser,
		Author:       gitConfig.Username,
		Email:        gitConfig.Email,
		SkeletonPath: skeletonPath,
	}

	return o, nil
}

func (o *CreateOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Author, "author", o.Author, "Project author")
	cmd.Flags().StringVar(&o.Email, "email", o.Email, "Project author's e-mail")
	cmd.Flags().StringVar(&o.ProjectName, "project-name", o.ProjectName, "Project name. Will be inferred from the output dir if not explicitly set")
	cmd.Flags().StringVar(&o.License, "license", o.License, "License to use for the project. If set this will automatically populate the LICENSE file")
	cmd.Flags().StringVar(&o.SkeletonPath, "skeleton-path", o.SkeletonPath, "Path to the skeleton. This can also be a git repository URL")
	cmd.Flags().StringVar(&o.GitHubUser, "github-user", o.GitHubUser, "GitHub username")
	cmd.Flags().StringVar(&o.GitHubRepo, "github-repo", o.GitHubRepo, "GitHub repo name (defaults to the project name)")
	cmd.Flags().BoolVar(&o.DryRun, "dry-run", o.DryRun, "Only print what would be done")
	cmd.Flags().BoolVar(&o.Force, "force", o.Force, "Forces overwrite of existing output directory")
}

func (o *CreateOptions) Complete(cmd *cobra.Command, args []string) error {
	absPath, err := filepath.Abs(args[0])
	if err != nil {
		return err
	}

	o.OutputDir = absPath

	if o.ProjectName == "" {
		o.ProjectName = filepath.Base(o.OutputDir)
	}

	if o.GitHubRepo == "" {
		o.GitHubRepo = o.ProjectName
	}

	o.GitHubRepoURL = fmt.Sprintf("https://github.com/%s/%s", o.GitHubUser, o.GitHubRepo)

	if o.SkeletonPath != "" {
		sp, err := filepath.Abs(o.SkeletonPath)
		if err != nil {
			return err
		}

		o.SkeletonPath = sp
	}

	if o.License != "" {
		o.LicenseInfo, err = license.Lookup(o.License)
		if errors.Is(err, license.ErrLicenseNotFound) {
			return fmt.Errorf("license %q not found, use the `list-licenses` command to get a list of available licenses", o.License)
		} else if err != nil {
			return err
		}
	}

	return nil
}

func (o *CreateOptions) Validate() error {
	_, err := os.Stat(o.OutputDir)
	if !os.IsNotExist(err) && !o.Force {
		return fmt.Errorf("output-dir %q already exists, add --force to overwrite", o.OutputDir)
	}

	if o.SkeletonPath == "" {
		return fmt.Errorf("--skeleton-path path must be provided")
	}

	if o.GitHubUser == "" {
		return fmt.Errorf("--github-user needs to be set")
	}

	return nil
}

func (o *CreateOptions) Run() error {
	templateVars := map[string]interface{}{
		"ProjectName": o.ProjectName,
		"Repo":        o.GitHubRepo,
		"RepoURL":     o.GitHubRepoURL,
		"User":        o.GitHubUser,
		"Author":      o.Author,
		"Email":       o.Email,
		"License":     o.LicenseInfo,
		"Custom":      map[string]interface{}{},
	}

	log.WithFields(log.Fields(templateVars)).Info("creating project with template vars")

	err := filepath.Walk(o.SkeletonPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(o.SkeletonPath, path)
		if err != nil {
			return err
		}

		outputPath := filepath.Join(o.OutputDir, relPath)

		if info.IsDir() {
			log.WithFields(log.Fields{"dir": outputPath}).Info("creating directory")

			if o.DryRun {
				return nil
			}

			if file.Exists(outputPath) {
				log.WithFields(log.Fields{"dir": outputPath}).Warn("directory already exists")
				return nil
			}

			return os.MkdirAll(outputPath, info.Mode())
		}

		ext := filepath.Ext(path)
		if ext != ".skel" {
			log.WithFields(log.Fields{"src": path, "dst": outputPath}).Info("copying file")

			if o.DryRun {
				return nil
			}

			if file.Exists(outputPath) {
				log.WithFields(log.Fields{"dst": outputPath}).Warn("file already exists")
			}

			return file.Copy(path, outputPath)
		}

		outputPath = strings.TrimRight(outputPath, ".skel")

		name := filepath.Base(path)

		tpl, err := template.New(name).ParseFiles(path)
		if err != nil {
			return err
		}

		var buf bytes.Buffer

		err = tpl.Execute(&buf, templateVars)
		if err != nil {
			return err
		}

		log.WithFields(log.Fields{"template": path, "dst": outputPath}).Info("writing template")

		if o.DryRun {
			return nil
		}

		if file.Exists(outputPath) {
			log.WithFields(log.Fields{"dst": outputPath}).Warn("file already exists")
		}

		return ioutil.WriteFile(outputPath, buf.Bytes(), info.Mode())
	})
	if err != nil {
		return err
	}

	err = o.initGitRepository()
	if err != nil {
		return err
	}

	return o.writeLicenseFile()
}

func (o *CreateOptions) initGitRepository() error {
	if o.DryRun {
		return nil
	}

	return git.EnsureInitialized(o.OutputDir)
}

func (o *CreateOptions) writeLicenseFile() error {
	if o.LicenseInfo == nil {
		return nil
	}

	log.Info("writing LICENSE file")

	if o.DryRun {
		return nil
	}

	body := o.LicenseInfo.Body
	body = strings.ReplaceAll(body, "[fullname]", fmt.Sprintf("%s <%s>", o.Author, o.Email))
	body = strings.ReplaceAll(body, "[year]", strconv.Itoa(time.Now().Year()))

	outputPath := filepath.Join(o.OutputDir, "LICENSE")

	return ioutil.WriteFile(outputPath, []byte(body), 0644)
}
