package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/kirsle/configdir"
	"github.com/martinohmann/kickoff/pkg/file"
	"github.com/spf13/cobra"
	"helm.sh/helm/pkg/strvals"
)

const (
	DefaultSkeleton = "default"

	SkeletonConfigFile = ".kickoff.yaml"
)

var (
	LocalDir = configdir.LocalConfig("kickoff")

	DefaultSkeletonsDir = filepath.Join(LocalDir, "skeletons")
	DefaultConfigPath   = filepath.Join(LocalDir, "config.yaml")
)

type Config struct {
	ProjectName  string
	License      string
	Author       *AuthorConfig
	Repository   *RepositoryConfig
	Skeleton     string
	SkeletonsDir string
	Custom       map[string]interface{}

	rawCustomValues []string
}

func NewDefaultConfig() *Config {
	return &Config{
		Author:     &AuthorConfig{},
		Repository: NewDefaultRepositoryConfig(),
		Skeleton:   DefaultSkeleton,
	}
}

func (c *Config) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&c.ProjectName, "project-name", c.ProjectName, "Name of the project. Will be inferred from the output dir if not explicitly set")
	cmd.Flags().StringVar(&c.License, "license", c.License, "License to use for the project. If set this will automatically populate the LICENSE file")
	cmd.Flags().StringVar(&c.Skeleton, "skeleton", c.Skeleton, "Name of the skeleton to create the project from")
	cmd.Flags().StringVar(&c.SkeletonsDir, "skeletons-dir", c.SkeletonsDir, fmt.Sprintf("Path to the skeletons directory. (defaults to %q if the directory exists)", DefaultSkeletonsDir))
	cmd.Flags().StringArrayVar(&c.rawCustomValues, "set", c.rawCustomValues, "Set custom config values of the form key1=value1,key2=value2,deeply.nested.key3=value")

	c.Author.AddFlags(cmd)
	c.Repository.AddFlags(cmd)
}

func (c *Config) SkeletonDir() string {
	return filepath.Join(c.SkeletonsDir, c.Skeleton)
}

func (c *Config) SkeletonConfigPath() string {
	return filepath.Join(c.SkeletonDir(), SkeletonConfigFile)
}

func (c *Config) Complete(outputDir string) (err error) {
	if c.ProjectName == "" {
		c.ProjectName = filepath.Base(outputDir)
	}

	if c.Skeleton == "" {
		c.Skeleton = DefaultSkeleton
	}

	if c.SkeletonsDir == "" && file.Exists(DefaultSkeletonsDir) {
		c.SkeletonsDir = DefaultSkeletonsDir
	}

	if c.SkeletonsDir != "" {
		c.SkeletonsDir, err = filepath.Abs(c.SkeletonsDir)
		if err != nil {
			return err
		}
	}

	if len(c.rawCustomValues) > 0 {
		for _, rawValues := range c.rawCustomValues {
			err = strvals.ParseInto(rawValues, c.Custom)
			if err != nil {
				return err
			}
		}
	}

	err = c.Repository.Complete(c.ProjectName)
	if err != nil {
		return err
	}

	return c.Author.Complete()
}

func (c *Config) Validate() error {
	if c.Skeleton == "" {
		return fmt.Errorf("--skeleton must be provided")
	}

	if c.SkeletonsDir == "" {
		return fmt.Errorf("--skeletons-dir must be provided")
	}

	return c.Repository.Validate()
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
