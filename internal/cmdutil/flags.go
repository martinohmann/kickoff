package cmdutil

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

func contains(haystack []string, needle string) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}

	return false
}

// AddOutputFlag adds the --output flag to cmd and binds it to p. The first
// value from allowedValues is set as the default. Panics if allowedValues is
// empty.
func AddOutputFlag(cmd *cobra.Command, p *string, allowedValues ...string) {
	if len(allowedValues) == 0 {
		panic("cmdutil.AddOutputFlag: allowedValues must not be empty")
	}

	allowedValuesMsg := fmt.Sprintf(`allowed values: "%s"`, strings.Join(allowedValues, `", "`))

	value := newValidatedStringValue(allowedValues[0], p, func(s string) error {
		for _, vv := range allowedValues {
			if vv == s {
				return nil
			}
		}

		return errors.New(allowedValuesMsg)
	})

	cmd.Flags().VarP(value, "output", "o", fmt.Sprintf(`Output format, %s`, allowedValuesMsg))
	cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return allowedValues, cobra.ShellCompDirectiveDefault
	})
}

// validatedStringValue is a pflag.Value that validates a string flag value
// while it is set.
type validatedStringValue struct {
	p        *string
	validate func(string) error
}

// newValidatedStringValue creates a *validatedStringValue with val as default
// value and binds it to p. The validator func validates the (new) flag value
// before it is set.
func newValidatedStringValue(val string, p *string, validator func(string) error) *validatedStringValue {
	*p = val
	return &validatedStringValue{
		p:        p,
		validate: validator,
	}
}

// String implements the pflag.Value interface.
func (f *validatedStringValue) String() string {
	return *f.p
}

// Set implements the pflag.Value interface.
func (f *validatedStringValue) Set(s string) error {
	if err := f.validate(s); err != nil {
		return err
	}

	*f.p = s
	return nil
}

// Type implements the pflag.Value interface.
func (f *validatedStringValue) Type() string {
	return "string"
}
