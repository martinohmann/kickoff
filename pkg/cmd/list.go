package cmd

import (
	"fmt"
	"io"

	"github.com/martinohmann/kickoff/pkg/config"
	"github.com/martinohmann/kickoff/pkg/kickoff"
	"github.com/spf13/cobra"
)

func NewListCmd() *cobra.Command {
	o := NewListOptions()

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available skeletons",
		Long:  "Lists all skeletons available in the skeletons-dir",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(""); err != nil {
				return err
			}

			return o.Run()
		},
	}

	o.Out = cmd.OutOrStdout()

	cmd.Flags().StringVar(&o.SkeletonsDir, "skeletons-dir", o.SkeletonsDir, fmt.Sprintf("Path to the skeletons directory. (defaults to %q if the directory exists)", config.DefaultSkeletonsDir))

	return cmd
}

type ListOptions struct {
	*config.Config
	Out io.Writer
}

func NewListOptions() *ListOptions {
	return &ListOptions{
		Config: config.NewDefaultConfig(),
	}
}

func (o *ListOptions) Run() error {
	skeletons, err := kickoff.FindSkeletons(o.SkeletonsDir)
	if err != nil {
		return err
	}

	fmt.Fprintf(o.Out, "Skeletons available in %s:\n\n", o.SkeletonsDir)

	for _, skeleton := range skeletons {
		fmt.Fprintln(o.Out, skeleton)
	}

	return nil
}
