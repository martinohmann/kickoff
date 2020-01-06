package config

import (
	"github.com/martinohmann/kickoff/pkg/skeleton"
	"github.com/spf13/cobra"
	"helm.sh/helm/pkg/strvals"
)

type SkeletonConfig struct {
	*skeleton.Config

	rawValues []string
}

func NewSkeletonConfig() *SkeletonConfig {
	return &SkeletonConfig{
		Config: &skeleton.Config{},
	}
}

func (c *SkeletonConfig) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&c.License, "license", c.License, "License to use for the project. If set this will automatically populate the LICENSE file")
	cmd.Flags().StringArrayVar(&c.rawValues, "set", c.rawValues, "Set custom values of the form key1=value1,key2=value2,deeply.nested.key3=value that are then made available to .skel templates")
}

func (c *SkeletonConfig) Complete() (err error) {
	if c.License == "" {
		c.License = "none" // sanitize as "none" and empty string are treated the same.
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
