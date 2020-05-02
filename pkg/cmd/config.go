package cmd

import (
	"github.com/spf13/cobra"
	"kickoff.run/pkg/cli"
	"kickoff.run/pkg/cmd/config"
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
