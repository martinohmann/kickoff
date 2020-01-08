package cmd

import (
	"github.com/apex/log"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	verbose := false

	cmd := &cobra.Command{
		Use:           "kickoff",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			if verbose {
				log.SetLevel(log.DebugLevel)
			}
		},
	}

	cmd.PersistentFlags().BoolVar(&verbose, "verbose", verbose, "Enable verbose log output")

	cmd.AddCommand(NewProjectCmd())
	cmd.AddCommand(NewLicenseCmd())
	cmd.AddCommand(NewSkeletonCmd())
	cmd.AddCommand(NewVersionCmd())

	return cmd
}
