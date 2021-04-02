package cmd

import (
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmd/cache"
	"github.com/spf13/cobra"
)

// NewCacheCmd creates a command which provides subcommands for interacting
// with the kickoff cache.
func NewCacheCmd(streams cli.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage kickoff cache",
	}

	cmd.AddCommand(cache.NewCleanCmd(streams))

	return cmd
}
