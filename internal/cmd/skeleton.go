package cmd

import (
	"github.com/martinohmann/kickoff/internal/cmd/skeleton"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewSkeletonCmd creates a new command which provides subcommands for creating
// and inspecting project skeletons.
func NewSkeletonCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "skeleton",
		Aliases: []string{"skel", "skeletons"},
		Short:   "Manage skeletons",
	}

	cmd.AddCommand(skeleton.NewCreateCmd(f))
	cmd.AddCommand(skeleton.NewListCmd(f))
	cmd.AddCommand(skeleton.NewShowCmd(f))
	cmd.AddCommand(skeleton.NewShowFileCmd(f))

	return cmd
}
