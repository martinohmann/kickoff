package cmdutil

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// AddRepositoryFlag adds the --repository flag to cmd and binds it to p.
func AddRepositoryFlag(cmd *cobra.Command, f *Factory, p *[]string) {
	cmd.Flags().StringSliceVarP(p, "repository", "r", *p,
		"Limit to the named repository. Can be specified multiple times to filter for multiple repositories.")
	cmd.RegisterFlagCompletionFunc("repository", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return RepositoryNames(f), cobra.ShellCompDirectiveDefault
	})
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
