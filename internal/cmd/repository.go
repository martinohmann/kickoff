package cmd

import (
	"github.com/martinohmann/kickoff/internal/cmd/repository"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewRepositoryCmd creates a new command which provides subcommands for
// managing, creating and inspecting skeleton repositories.
func NewRepositoryCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "repository",
		Aliases: []string{"repo", "repos", "repositories"},
		Short:   "Manage repositories",
	}

	cmd.AddCommand(repository.NewAddCmd(f))
	cmd.AddCommand(repository.NewCreateCmd(f))
	cmd.AddCommand(repository.NewListCmd(f))
	cmd.AddCommand(repository.NewRemoveCmd(f))

	return cmd
}
