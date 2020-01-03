package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/apex/log"
	"github.com/spf13/cobra"
	gitconfig "github.com/tcnksm/go-gitconfig"
)

const skeletonBase = "_skeleton"

func newCreateCommand() *cobra.Command {
	o, err := NewDefaultCreateOptions()
	if err != nil {
		panic(err)
	}

	cmd := &cobra.Command{
		Use:   "create <project-name> [<output-dir>]",
		Short: "Create golang project skeletons",
		Args:  cobra.MinimumNArgs(1),
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

	GitHubUser    string
	GitHubRepo    string
	GitHubRepoURL string

	SkeletonPath string
	OutputDir    string

	DryRun bool
	Force  bool
}

func NewDefaultCreateOptions() (*CreateOptions, error) {
	gitUser, err := gitconfig.Username()
	if err != nil {
		u, err := user.Current()
		if err != nil {
			return nil, err
		}

		if u.Name != "" {
			gitUser = u.Name
		} else {
			gitUser = u.Username
		}
	}

	githubUser, err := gitconfig.GithubUser()
	if err != nil {
		githubUser = gitUser
	}

	gitEmail, _ := gitconfig.Email()

	var skeletonPath string
	_, err = os.Stat(skeletonBase)
	if err == nil {
		skeletonPath, err = filepath.Abs(skeletonBase)
		if err != nil {
			return nil, err
		}
	}

	o := &CreateOptions{
		GitHubUser:   githubUser,
		Author:       gitUser,
		Email:        gitEmail,
		SkeletonPath: skeletonPath,
	}

	return o, nil
}

func (o *CreateOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Author, "author", o.Author, "Project author")
	cmd.Flags().StringVar(&o.Email, "email", o.Email, "Project author's e-mail")
	cmd.Flags().StringVar(&o.SkeletonPath, "skeleton-path", o.SkeletonPath, "Path to the skeleton. This can also be a git repository URL")
	cmd.Flags().StringVar(&o.GitHubUser, "github-user", o.GitHubUser, "GitHub username")
	cmd.Flags().StringVar(&o.GitHubRepo, "github-repo", o.GitHubRepo, "GitHub repo name (defaults to the project name)")
	cmd.Flags().BoolVar(&o.DryRun, "dry-run", o.DryRun, "Only print what would be done")
	cmd.Flags().BoolVar(&o.Force, "force", o.Force, "Forces overwrite of existing output directory")
}

func (o *CreateOptions) Complete(cmd *cobra.Command, args []string) error {
	o.ProjectName = args[0]

	if len(args) > 1 {
		absPath, err := filepath.Abs(args[1])
		if err != nil {
			return err
		}

		o.OutputDir = absPath
	} else {
		pwd, err := os.Getwd()
		if err != nil {
			return err
		}

		o.OutputDir = filepath.Join(pwd, o.ProjectName)
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
	}

	log.WithFields(log.Fields(templateVars)).Info("creating project with template vars")

	return filepath.Walk(o.SkeletonPath, func(path string, info os.FileInfo, err error) error {
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

			if fileExists(outputPath) {
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

			if fileExists(outputPath) {
				log.WithFields(log.Fields{"dst": outputPath}).Warn("file already exists")
			}

			return copyFile(path, outputPath)
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

		if fileExists(outputPath) {
			log.WithFields(log.Fields{"dst": outputPath}).Warn("file already exists")
		}

		return ioutil.WriteFile(outputPath, buf.Bytes(), info.Mode())
	})
}

func copyFile(srcPath, dstPath string) error {
	fileInfo, err := os.Stat(srcPath)
	if err != nil {
		return err
	}

	if !fileInfo.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", srcPath)
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(dstPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileInfo.Mode())
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	return err
}

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
