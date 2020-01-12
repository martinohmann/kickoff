package cmdutil

import (
	"fmt"

	"github.com/martinohmann/kickoff/pkg/config"
	"github.com/spf13/cobra"
)

// AddConfigFlag adds the --config flag to cmd and binds it to val.
func AddConfigFlag(cmd *cobra.Command, val *string) {
	cmd.Flags().StringVar(val, "config", *val, fmt.Sprintf("Path to config file (defaults to %q if the file exists)", config.DefaultConfigPath))
}

// AddForceFlag adds the --force flag to cmd and binds it to val.
func AddForceFlag(cmd *cobra.Command, val *bool) {
	cmd.Flags().BoolVar(val, "force", *val, "Forces overwrite of existing output directory")
}
