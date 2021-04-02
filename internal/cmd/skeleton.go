package cmd

import (
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmd/skeleton"
	"github.com/spf13/cobra"
)

// NewSkeletonCmd creates a new command which provides subcommands for creating
// and inspecting project skeletons.
func NewSkeletonCmd(streams cli.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "skeleton",
		Aliases: []string{"skel", "skeletons"},
		Short:   "Manage skeletons",
	}

	cmd.AddCommand(skeleton.NewCreateCmd(streams))
	cmd.AddCommand(skeleton.NewListCmd(streams))
	cmd.AddCommand(skeleton.NewShowCmd(streams))

	return cmd
}
