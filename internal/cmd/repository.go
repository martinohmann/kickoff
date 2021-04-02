package cmd

import (
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmd/repository"
	"github.com/spf13/cobra"
)

// NewRepositoryCmd creates a new command which provides subcommands for
// managing, creating and inspecting skeleton repositories.
func NewRepositoryCmd(streams cli.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "repository",
		Aliases: []string{"repo", "repos", "repositories"},
		Short:   "Manage repositories",
	}

	cmd.AddCommand(repository.NewAddCmd(streams))
	cmd.AddCommand(repository.NewCreateCmd(streams))
	cmd.AddCommand(repository.NewListCmd(streams))
	cmd.AddCommand(repository.NewRemoveCmd(streams))

	return cmd
}
