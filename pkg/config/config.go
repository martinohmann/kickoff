package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/martinohmann/skeleton-go/pkg/file"
	"github.com/martinohmann/skeleton-go/pkg/git"
	"github.com/spf13/cobra"
)

const (
	DefaultSkeletonPath = ".skeleton-go"
	DefaultConfigPath   = ".skeleton-go.yaml"
)

type Config struct {
	ProjectName string
	License     string
	Author      AuthorConfig
	Repository  RepositoryConfig
	Skeleton    SkeletonConfig
	Custom      map[string]interface{}
}

func (c *Config) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&c.Author.Fullname, "author-fullname", c.Author.Fullname, "Project author's fullname")
	cmd.Flags().StringVar(&c.Author.Email, "author-email", c.Author.Email, "Project author's e-mail")
	cmd.Flags().StringVar(&c.ProjectName, "project-name", c.ProjectName, "Name of the project. Will be inferred from the output dir if not explicitly set")
	cmd.Flags().StringVar(&c.License, "license", c.License, "License to use for the project. If set this will automatically populate the LICENSE file")
	cmd.Flags().StringVar(&c.Skeleton.Path, "skeleton-path", c.Skeleton.Path, fmt.Sprintf("Path to the skeleton. (defaults to %q if the directory exists)", DefaultSkeletonPath))
	cmd.Flags().StringVar(&c.Repository.User, "repository-user", c.Repository.User, "Repository username")
	cmd.Flags().StringVar(&c.Repository.Name, "repository-name", c.Repository.Name, "Repository name (defaults to the project name)")
}

func (c *Config) Complete(outputDir string) (err error) {
	gitConfig := git.GlobalConfig()

	if c.Author.Fullname == "" {
		c.Author.Fullname = gitConfig.Username
	}

	if c.Author.Email == "" {
		c.Author.Email = gitConfig.Email
	}

	if c.Repository.User == "" {
		c.Repository.User = gitConfig.GitHubUser
	}

	if c.ProjectName == "" {
		c.ProjectName = filepath.Base(outputDir)
	}

	if c.Repository.Name == "" {
		c.Repository.Name = c.ProjectName
	}

	if c.Skeleton.Path == "" && file.Exists(DefaultSkeletonPath) {
		c.Skeleton.Path = DefaultSkeletonPath
	}

	if c.Skeleton.Path != "" {
		c.Skeleton.Path, err = filepath.Abs(c.Skeleton.Path)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) Validate() error {
	if c.Skeleton.Path == "" {
		return fmt.Errorf("--skeleton-path path must be provided")
	}

	if c.Repository.User == "" {
		return fmt.Errorf("--repository-user needs to be set")
	}

	return nil
}

type AuthorConfig struct {
	Fullname string
	Email    string
}

func (c AuthorConfig) String() string {
	return fmt.Sprintf("%s <%s>", c.Fullname, c.Email)
}

type RepositoryConfig struct {
	User string
	Name string
}

type SkeletonConfig struct {
	Path string
}

func Load(filePath string) (*Config, error) {
	buf, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config Config

	err = yaml.Unmarshal(buf, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
