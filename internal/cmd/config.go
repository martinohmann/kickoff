package cmd

import (
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmd/config"
	"github.com/spf13/cobra"
)

// NewConfigCmd creates a new command which provides subcommands for
// manipulating and inspecting the kickoff config.
func NewConfigCmd(streams cli.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config",
		Aliases: []string{"conf"},
		Short:   "Manage kickoff config",
	}

	cmd.AddCommand(config.NewEditCmd(streams))
	cmd.AddCommand(config.NewShowCmd(streams))

	return cmd
}
