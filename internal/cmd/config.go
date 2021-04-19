package cmd

import (
	"github.com/martinohmann/kickoff/internal/cmd/config"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewConfigCmd creates a new command which provides subcommands for
// manipulating and inspecting the kickoff config.
func NewConfigCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config",
		Aliases: []string{"conf"},
		Short:   "Manage kickoff config",
	}

	cmd.AddCommand(config.NewEditCmd(f))
	cmd.AddCommand(config.NewShowCmd(f))

	return cmd
}
