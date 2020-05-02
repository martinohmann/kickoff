package cmd

import (
	"github.com/spf13/cobra"
	"kickoff.run/pkg/cli"
	"kickoff.run/pkg/cmd/gitignore"
)

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
