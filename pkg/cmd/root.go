package cmd

import (
	"github.com/apex/log"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	debug := false

	cmd := &cobra.Command{
		Use:           "skeleton-go",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			if debug {
				log.SetLevel(log.DebugLevel)
			}
		},
	}

	cmd.PersistentFlags().BoolVar(&debug, "debug", debug, "Enable debug log")

	return cmd
}
