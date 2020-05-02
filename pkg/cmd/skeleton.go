package cmd

import (
	"github.com/spf13/cobra"
	"kickoff.run/pkg/cli"
	"kickoff.run/pkg/cmd/skeleton"
)

func NewSkeletonCmd(streams cli.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "skeleton",
		Aliases: []string{"skel", "skeletons"},
		Short:   "Manage skeletons",
	}

	cmd.AddCommand(skeleton.NewCreateCmd())
	cmd.AddCommand(skeleton.NewListCmd(streams))
	cmd.AddCommand(skeleton.NewShowCmd(streams))

	return cmd
}
