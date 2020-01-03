package cmd

import (
	"os"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	debug := false

	cmd := &cobra.Command{
		Use:           "skeleton-go",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			logHandler := cli.New(os.Stdout)
			logHandler.Padding = 0

			log.SetHandler(logHandler)

			if debug {
				log.SetLevel(log.DebugLevel)
			}
		},
	}

	cmd.PersistentFlags().BoolVar(&debug, "debug", debug, "Enable debug log")

	return cmd
}
