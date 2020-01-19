package cmd

import (
	"github.com/apex/log"
	"github.com/martinohmann/kickoff/pkg/cli"
	"github.com/spf13/cobra"
)

func NewRootCmd(streams cli.IOStreams) *cobra.Command {
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

	cmd.AddCommand(NewConfigCmd(streams))
	cmd.AddCommand(NewGitignoreCmd(streams))
	cmd.AddCommand(NewInitCmd(streams))
	cmd.AddCommand(NewProjectCmd())
	cmd.AddCommand(NewLicenseCmd(streams))
	cmd.AddCommand(NewRepositoryCmd(streams))
	cmd.AddCommand(NewSkeletonCmd(streams))
	cmd.AddCommand(NewVersionCmd(streams))

	return cmd
}
