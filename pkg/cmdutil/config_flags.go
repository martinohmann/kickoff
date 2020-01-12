package cmdutil

import (
	"github.com/apex/log"
	"github.com/martinohmann/kickoff/pkg/config"
	"github.com/martinohmann/kickoff/pkg/file"
	"github.com/spf13/cobra"
)

// ConfigFlags provide a flag for configuring the path to the kickoff config
// file. Can be used to automatically populate the kickoff config with defaults
// and optionally override them if the user passed a different config path via
// the CLI flag.
type ConfigFlags struct {
	config.Config

	ConfigPath string
}

// AddFlags adds flags for configuring the config file location to cmd.
func (f *ConfigFlags) AddFlags(cmd *cobra.Command) {
	AddConfigFlag(cmd, &f.ConfigPath)
	cmd.Flags().StringToStringVar(&f.Repositories, "repositories", f.Repositories, "Skeleton repositories of the form name1=url1,name2=url2. The repository urls can be a local path or a remote git repository.")
}

// Complete completes the embedded kickoff configuration using the provided
// defaultProjectName. It will load the config file from the path provided by
// the user and merge it into the configuration and apply configuration
// defaults to unset fields. Returns an error if the config file does not
// exist, could not be read or contains invalid configuration. If the user did
// not provide any config file path, the default config file will be loaded
// instead, if it exists.
func (f *ConfigFlags) Complete(defaultProjectName string) (err error) {
	if f.ConfigPath == "" && file.Exists(config.DefaultConfigPath) {
		f.ConfigPath = config.DefaultConfigPath
	}

	if f.ConfigPath != "" {
		log.WithField("path", f.ConfigPath).Debugf("loading config file")

		err = f.Config.MergeFromFile(f.ConfigPath)
		if err != nil {
			return err
		}
	}

	f.ApplyDefaults(defaultProjectName)

	return nil
}
