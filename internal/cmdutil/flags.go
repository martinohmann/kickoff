package cmdutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/martinohmann/kickoff/internal/file"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/spf13/cobra"
)

// AddConfigFlag adds the --config flag to cmd and binds it to val.
func AddConfigFlag(cmd *cobra.Command, val *string) {
	cmd.Flags().StringVarP(val, "config", "c", *val, fmt.Sprintf("Path to config file (defaults to %q if the file exists)", kickoff.DefaultConfigPath))
	cmd.MarkFlagFilename("config", "yaml", "yml")
}

// ConfigFlags provide a flag for configuring the path to the kickoff config
// file. Can be used to automatically populate the kickoff config with defaults
// and optionally override them if the user passed a different config path via
// the CLI flag.
type ConfigFlags struct {
	kickoff.Config

	ConfigPath string

	repositories []string
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
		f.ConfigPath = kickoff.DefaultConfigPath

		if configPath := os.Getenv("KICKOFF_CONFIG"); configPath != "" {
			f.ConfigPath = configPath
		}
	}

	f.ConfigPath, err = filepath.Abs(f.ConfigPath)
	if err != nil {
		return err
	}

	// We allow the default config file to be absent and simply skip loading it
	// if it does not exist. For all other config file paths this will result
	// in a load error. This is a convenience feature to keep kickoff usable
	// without a config file.
	skipLoad := f.ConfigPath == kickoff.DefaultConfigPath && !file.Exists(f.ConfigPath)

	if !skipLoad {
		if err := f.Config.MergeFromFile(f.ConfigPath); err != nil {
			return err
		}
	}

	f.ApplyDefaults()

	if len(f.repositories) > 0 {
		// Ensure that the repos provided by the user are configured.
		for _, name := range f.repositories {
			if _, ok := f.Repositories[name]; !ok {
				return RepositoryNotConfiguredError(name)
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
	Output      string
	ValidValues []string
}

// NewOutputFlag creates a new OutputFlag with a list of valid values. If
// empty, validValues defaults to [json, yaml].
func NewOutputFlag(validValues ...string) OutputFlag {
	return OutputFlag{ValidValues: validValues}
}

// AddFlag adds the flag for configuring output format to cmd.
func (f *OutputFlag) AddFlag(cmd *cobra.Command) {
	if len(f.ValidValues) == 0 {
		f.ValidValues = []string{"json", "yaml"}
	}

	cmd.Flags().StringVarP(&f.Output, "output", "o", f.Output, "Output format")
	cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return f.ValidValues, cobra.ShellCompDirectiveDefault
	})
}

// Validate validates the output format and returns an error if the user
// provided an invalid value.
func (f *OutputFlag) Validate() error {
	if f.Output == "" {
		return nil
	}

	for _, v := range f.ValidValues {
		if v == f.Output {
			return nil
		}
	}

	sort.Strings(f.ValidValues)

	return fmt.Errorf(`--output must be one of: '%s'`, strings.Join(f.ValidValues, `', '`))
}

// TimeoutFlag configure the timeout for operations that cross API boundaries,
// such as http requests to third-party integrations.
type TimeoutFlag struct {
	Timeout time.Duration
}

// NewDefaultTimeoutFlag creates a new TimeoutFlag which uses the
// DefaultTimeout is not overridded.
func NewDefaultTimeoutFlag() TimeoutFlag {
	return TimeoutFlag{Timeout: 20 * time.Second}
}

// AddFlag adds the timeout flag to cmd.
func (f *TimeoutFlag) AddFlag(cmd *cobra.Command) {
	cmd.Flags().DurationVar(&f.Timeout, "timeout", f.Timeout, "Timeout for remote operations. Zero or less means that there is no timeout.")
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
