package cmd

import (
	"github.com/apex/log"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	verbose := false

	cmd := &cobra.Command{
		Use:           "skeleton-go",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			if verbose {
				log.SetLevel(log.DebugLevel)
			}
		},
	}

	cmd.PersistentFlags().BoolVar(&verbose, "verbose", verbose, "Enable verbose log output")

	return cmd
}
