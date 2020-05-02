package cmd

import (
	"github.com/spf13/cobra"
	"kickoff.run/pkg/cmd/project"
)

func NewProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "project",
		Aliases: []string{"proj", "projects"},
		Short:   "Manage projects",
	}

	cmd.AddCommand(project.NewCreateCmd())

	return cmd
}
