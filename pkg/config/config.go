package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/apex/log"
	"github.com/ghodss/yaml"
	"github.com/imdario/mergo"
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
	LocalConfigDir      = configdir.LocalConfig("kickoff")
	DefaultSkeletonsDir = filepath.Join(LocalConfigDir, "skeletons")
	DefaultConfigPath   = filepath.Join(LocalConfigDir, "config.yaml")
)

type Config struct {
	ProjectName  string                 `json:"projectName"`
	License      string                 `json:"license"`
	Author       *AuthorConfig          `json:"author"`
	Repository   *RepositoryConfig      `json:"repository"`
	Skeleton     string                 `json:"skeleton"`
	SkeletonsDir string                 `json:"skeletonsDir"`
	CustomValues map[string]interface{} `json:"customValues"`

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
	cmd.Flags().StringArrayVar(&c.rawCustomValues, "set-custom", c.rawCustomValues, "Set custom config values of the form key1=value1,key2=value2,deeply.nested.key3=value")

	c.Author.AddFlags(cmd)
	c.Repository.AddFlags(cmd)
}

func (c *Config) SkeletonDir() string {
	return filepath.Join(c.SkeletonsDir, c.Skeleton)
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
			err = strvals.ParseInto(rawValues, c.CustomValues)
			if err != nil {
				return err
			}
		}
	}

	skeletonDir := c.SkeletonDir()

	ok, err := file.IsDirectory(skeletonDir)
	if err != nil {
		return fmt.Errorf("failed to stat skeleton directory: %v", err)
	}

	if !ok {
		return fmt.Errorf("invalid skeleton: %s is not a directory", skeletonDir)
	}

	skeletonConfigPath := filepath.Join(skeletonDir, SkeletonConfigFile)

	if file.Exists(skeletonConfigPath) {
		log.WithField("skeleton", c.SkeletonDir()).Debugf("found %s, merging config values", SkeletonConfigFile)

		err = c.MergeFromFile(skeletonConfigPath)
		if err != nil {
			return err
		}
	}

	if c.License == "none" {
		c.License = "" // sanitize as "none" and empty string are treated the same
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

// MergeFromFile loads the config from path and merges it into c. Returns any
// errors that may occur during loading or merging.
func (c *Config) MergeFromFile(path string) error {
	config, err := Load(path)
	if err != nil {
		return err
	}

	return mergo.Merge(c, config)
}

// Load loads the config from path and returns it.
func Load(path string) (*Config, error) {
	buf, err := ioutil.ReadFile(path)
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
