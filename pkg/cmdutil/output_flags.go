package cmdutil

import (
	"github.com/spf13/cobra"
)

// OutputFlags manage and validate flags related to output format.
type OutputFlags struct {
	Output string
}

// AddFlags adds flags for configuring output format to cmd.
func (f *OutputFlags) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.Output, "output", f.Output, "Output format")
}

// Validate validates the output format and returns an error if the user
// provided an invalid value.
func (f *OutputFlags) Validate() error {
	if f.Output != "" && f.Output != "yaml" && f.Output != "json" {
		return ErrInvalidOutputFormat
	}

	return nil
}
