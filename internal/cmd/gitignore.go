package cmd

import (
	"github.com/martinohmann/kickoff/internal/cmd/gitignore"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewGitignoreCmd creates a new command which provides subcommands for
// inspecting gitignore templates available via the GitHub Gitignores API.
func NewGitignoreCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gitignore",
		Aliases: []string{"gi", "gitignores"},
		Short:   "Inspect gitignore templates",
	}

	cmd.AddCommand(gitignore.NewListCmd(f))
	cmd.AddCommand(gitignore.NewShowCmd(f))

	return cmd
}
