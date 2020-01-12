package cmd

import (
	"github.com/martinohmann/kickoff/pkg/cli"
	"github.com/martinohmann/kickoff/pkg/cmd/repository"
	"github.com/spf13/cobra"
)

func NewRepositoryCmd(streams cli.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "repository",
		Aliases: []string{"repo", "repos", "repositories"},
		Short:   "Manage repositories",
	}

	cmd.AddCommand(repository.NewListCmd(streams))

	return cmd
}
