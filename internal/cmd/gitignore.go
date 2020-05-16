package cmd

import (
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmd/gitignore"
	"github.com/spf13/cobra"
)

// NewGitignoreCmd creates a new command which provides subcommands for
// inspecting gitignore templates provided by gitignore.io.
func NewGitignoreCmd(streams cli.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gitignore",
		Aliases: []string{"gi", "gitignores"},
		Short:   "Inspect gitignore templates",
	}

	cmd.AddCommand(gitignore.NewListCmd(streams))
	cmd.AddCommand(gitignore.NewShowCmd(streams))

	return cmd
}
