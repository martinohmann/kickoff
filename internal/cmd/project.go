package cmd

import (
	"github.com/martinohmann/kickoff/internal/cmd/project"
	"github.com/spf13/cobra"
)

// NewProjectCmd creates a new command which provides subcommands for working
// with projects.
func NewProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "project",
		Aliases: []string{"proj", "projects"},
		Short:   "Manage projects",
	}

	cmd.AddCommand(project.NewCreateCmd())

	return cmd
}
