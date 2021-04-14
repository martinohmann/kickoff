package cmdutil

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ExactNonEmptyArgs returns an error if there are not exactly n args or if any
// of them is an empty string.
func ExactNonEmptyArgs(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != n {
			return fmt.Errorf("accepts %d non-empty arg(s), received %d", n, len(args))
		}

		for i, arg := range args {
			if arg == "" {
				return fmt.Errorf("accepts %d non-empty arg(s), received empty arg at position %d", n, i+1)
			}
		}

		return nil
	}
}
