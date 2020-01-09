package cmd

import (
	"github.com/martinohmann/kickoff/pkg/cli"
	"github.com/martinohmann/kickoff/pkg/cmd/skeleton"
	"github.com/spf13/cobra"
)

func NewSkeletonCmd(streams cli.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "skeleton",
		Aliases: []string{"skel", "skeletons"},
		Short:   "Manage skeletons",
	}

	cmd.AddCommand(skeleton.NewInitCmd())
	cmd.AddCommand(skeleton.NewListCmd(streams))
	cmd.AddCommand(skeleton.NewShowCmd(streams))

	return cmd
}
