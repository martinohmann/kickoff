package cmd

import (
	"github.com/apex/log"
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root command for kickoff.
func NewRootCmd(streams cli.IOStreams) *cobra.Command {
	logLevel := "warn"

	cmd := &cobra.Command{
		Use:           "kickoff",
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			lvl, err := log.ParseLevel(logLevel)
			if err != nil {
				return err
			}

			log.SetLevel(lvl)

			// We silence usage output here instead of doing so while
			// initializing the struct above because we want to print the usage
			// if the user actually misused the CLI (e.g. missing arguments,
			// wrong flags), but we do not want to show the usage on errors
			// that occurred when the CLI arguments where actually valid (e.g.
			// user queried for a skeleton that does not exist). Since
			// PersistentPreRun is called after argument parsing happened, we
			// silence the usage here for all subsequent errors.
			//
			// Also see the following issue:
			// https://github.com/spf13/cobra/issues/340#issuecomment-378726225
			cmd.SilenceUsage = true

			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&logLevel, "log-level", logLevel, "Level for stderr log output")

	cmd.AddCommand(NewCacheCmd(streams))
	cmd.AddCommand(NewCompletionCmd(streams))
	cmd.AddCommand(NewConfigCmd(streams))
	cmd.AddCommand(NewGitignoreCmd(streams))
	cmd.AddCommand(NewInitCmd(streams))
	cmd.AddCommand(NewLicenseCmd(streams))
	cmd.AddCommand(NewProjectCmd(streams))
	cmd.AddCommand(NewRepositoryCmd(streams))
	cmd.AddCommand(NewSkeletonCmd(streams))
	cmd.AddCommand(NewVersionCmd(streams))

	return cmd
}
