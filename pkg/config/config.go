package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/imdario/mergo"
	"github.com/kirsle/configdir"
	"github.com/martinohmann/kickoff/pkg/file"
	"github.com/spf13/cobra"
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
	ProjectName  string            `json:"projectName"`
	From         string            `json:"from"`
	SkeletonsDir string            `json:"skeletonsDir"`
	Author       *AuthorConfig     `json:"author"`
	Repository   *RepositoryConfig `json:"repository"`
	Skeleton     *SkeletonConfig   `json:"skeleton"`
}

func NewDefaultConfig() *Config {
	return &Config{
		Author:     &AuthorConfig{},
		Repository: NewDefaultRepositoryConfig(),
		Skeleton:   NewSkeletonConfig(),
		From:       DefaultSkeleton,
	}
}

func (c *Config) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&c.ProjectName, "project-name", c.ProjectName, "Name of the project. Will be inferred from the output dir if not explicitly set")
	cmd.Flags().StringVar(&c.From, "from", c.From, "Name of the skeleton to create the project from")
	cmd.Flags().StringVar(&c.SkeletonsDir, "skeletons-dir", c.SkeletonsDir, fmt.Sprintf("Path to the skeletons directory. (defaults to %q if the directory exists)", DefaultSkeletonsDir))

	c.Author.AddFlags(cmd)
	c.Repository.AddFlags(cmd)
	c.Skeleton.AddFlags(cmd)
}

func (c *Config) SkeletonDir() string {
	return filepath.Join(c.SkeletonsDir, c.From)
}

func (c *Config) Complete(outputDir string) (err error) {
	if c.ProjectName == "" {
		c.ProjectName = filepath.Base(outputDir)
	}

	if c.From == "" {
		c.From = DefaultSkeleton
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

	err = c.Repository.Complete(c.ProjectName)
	if err != nil {
		return err
	}

	err = c.Skeleton.Complete()
	if err != nil {
		return err
	}

	return c.Author.Complete()
}

func (c *Config) Validate() error {
	if c.From == "" {
		return fmt.Errorf("--from must be provided")
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
