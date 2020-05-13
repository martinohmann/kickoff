package cmd

import (
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmd/config"
	"github.com/spf13/cobra"
)

func NewConfigCmd(streams cli.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config",
		Aliases: []string{"conf"},
		Short:   "Manage kickoff config",
	}

	cmd.AddCommand(config.NewEditCmd())
	cmd.AddCommand(config.NewShowCmd(streams))

	return cmd
}
