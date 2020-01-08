package cmd

import (
	"github.com/martinohmann/kickoff/pkg/cmd/skeleton"
	"github.com/spf13/cobra"
)

func NewSkeletonCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "skeleton",
		Aliases: []string{"skel", "skeletons"},
		Short:   "Manage skeletons",
	}

	cmd.AddCommand(skeleton.NewListCmd())

	return cmd
}
