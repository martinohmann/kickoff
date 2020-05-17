package cmdutil

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/apex/log"
	"github.com/martinohmann/kickoff/internal/config"
	"github.com/martinohmann/kickoff/internal/file"
	"github.com/spf13/cobra"
)

const (
	// DefaultTimeout is the default timeout for operations that cross API
	// boundaries.
	DefaultTimeout = 20 * time.Second
)

// AddConfigFlag adds the --config flag to cmd and binds it to val.
func AddConfigFlag(cmd *cobra.Command, val *string) {
	cmd.Flags().StringVarP(val, "config", "c", *val, fmt.Sprintf("Path to config file (defaults to %q if the file exists)", config.DefaultConfigPath))
	cmd.MarkFlagFilename("config")
}

// AddForceFlag adds the --force flag to cmd and binds it to val.
func AddForceFlag(cmd *cobra.Command, val *bool) {
	cmd.Flags().BoolVar(val, "force", *val, "Forces writing into existing output directory")
}

// AddOverwriteFlag adds the --overwrite flag to cmd and binds it to val.
func AddOverwriteFlag(cmd *cobra.Command, val *bool) {
	cmd.Flags().BoolVar(val, "overwrite", *val, "Overwrite files that are already present in output directory")
}

// ConfigFlags provide a flag for configuring the path to the kickoff config
// file. Can be used to automatically populate the kickoff config with defaults
// and optionally override them if the user passed a different config path via
// the CLI flag.
type ConfigFlags struct {
	config.Config

	ConfigPath         string
	allowMissingConfig bool
	repositories       []string
}

// AllowMissingConfig allows the config file at ConfigPath to be absent. A
// nonexistent config file will not cause errors but is simply ignored. This is
// useful for initialization commands to be able to specify an alternative
// config file which may not exist yet.
func (f *ConfigFlags) AllowMissingConfig() {
	f.allowMissingConfig = true
}

// AddFlags adds flags for configuring the config file location to cmd.
func (f *ConfigFlags) AddFlags(cmd *cobra.Command) {
	AddConfigFlag(cmd, &f.ConfigPath)
	cmd.Flags().StringSliceVarP(&f.repositories, "repository", "r", f.repositories, "Limit to the named repository. Can be specified multiple times to filter for multiple repositories.")
}

// Complete completes the embedded kickoff configuration. It will load the
// config file from the path provided by the user and merge it into the
// configuration and apply configuration defaults to unset fields. Returns an
// error if the config file does not exist, could not be read or contains
// invalid configuration. If the user did not provide any config file path, the
// default config file will be loaded instead, if it exists.
func (f *ConfigFlags) Complete() (err error) {
	if f.ConfigPath == "" {
		if configPath := os.Getenv("KICKOFF_CONFIG"); configPath != "" {
			f.ConfigPath = configPath
		} else if file.Exists(config.DefaultConfigPath) {
			f.ConfigPath = config.DefaultConfigPath
		}
	}

	loadConfig := f.ConfigPath != "" && (!f.allowMissingConfig || file.Exists(f.ConfigPath))

	if loadConfig {
		log.WithField("path", f.ConfigPath).Debugf("loading config file")

		err = f.Config.MergeFromFile(f.ConfigPath)
		if err != nil {
			return err
		}
	}

	f.ApplyDefaults()

	if len(f.repositories) > 0 {
		// Ensure that the repos provided by the user are configured.
		for _, name := range f.repositories {
			_, ok := f.Repositories[name]
			if !ok {
				return fmt.Errorf("repository %q not configured", name)
			}
		}

		// Filter out repositories that do not match.
		for name := range f.Repositories {
			if !contains(f.repositories, name) {
				delete(f.Repositories, name)
			}
		}
	}

	return nil
}

// OutputFlag manage and validate a flag related to output format.
type OutputFlag struct {
	Output string
}

// AddFlag adds the flag for configuring output format to cmd.
func (f *OutputFlag) AddFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&f.Output, "output", "o", f.Output, "Output format")
}

// Validate validates the output format and returns an error if the user
// provided an invalid value.
func (f *OutputFlag) Validate() error {
	if f.Output != "" && f.Output != "yaml" && f.Output != "json" {
		return ErrInvalidOutputFormat
	}

	return nil
}

// TimeoutFlag configure the timeout for operations that cross API boundaries,
// such as http requests to third-party integrations.
type TimeoutFlag struct {
	Timeout time.Duration
}

// NewDefaultTimeoutFlag creates a new TimeoutFlag which uses the
// DefaultTimeout is not overridded.
func NewDefaultTimeoutFlag() TimeoutFlag {
	return TimeoutFlag{
		Timeout: DefaultTimeout,
	}
}

// AddFlag adds the timeout flag to cmd.
func (f *TimeoutFlag) AddFlag(cmd *cobra.Command) {
	cmd.Flags().DurationVar(&f.Timeout, "timeout", f.Timeout, "Timeout for http requests. Zero or less means that there is no timeout.")
}

// Context returns a context with the timeout set and a cancel func to cancel
// the context. If the timeout less or equal to zero, a normal background
// context is returned.
func (f *TimeoutFlag) Context() (context.Context, func()) {
	ctx := context.Background()

	if f.Timeout <= 0 {
		return context.WithCancel(ctx)
	}

	return context.WithTimeout(ctx, f.Timeout)
}

func contains(haystack []string, needle string) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}

	return false
}
