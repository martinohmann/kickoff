package config

import (
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/imdario/mergo"
	"github.com/spf13/cobra"
	"helm.sh/helm/pkg/strvals"
)

type SkeletonConfig struct {
	License string                 `json:"license"`
	Values  map[string]interface{} `json:"values"`

	rawValues []string
}

func (c *SkeletonConfig) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&c.License, "license", c.License, "License to use for the project. If set this will automatically populate the LICENSE file")
	cmd.Flags().StringArrayVar(&c.rawValues, "set", c.rawValues, "Set custom values of the form key1=value1,key2=value2,deeply.nested.key3=value that are then made available to .skel templates")
}

func (c *SkeletonConfig) Complete() (err error) {
	if c.License == "none" {
		c.License = "" // sanitize as "none" and empty string are treated the same.
	}

	if len(c.rawValues) == 0 {
		return nil
	}

	for _, rawValues := range c.rawValues {
		err = strvals.ParseInto(rawValues, c.Values)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *SkeletonConfig) MergeFromFile(path string) error {
	config, err := LoadSkeleton(path)
	if err != nil {
		return err
	}

	return mergo.Merge(c, config)
}

// LoadSkeleton loads the skeleton config from path and returns it.
func LoadSkeleton(path string) (*SkeletonConfig, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config SkeletonConfig

	err = yaml.Unmarshal(buf, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
