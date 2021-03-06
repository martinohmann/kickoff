package cache

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/kickoff"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewCleanCmd create a command for cleaning the kickoff cache.
func NewCleanCmd(streams cli.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Cleans the kickoff cache",
		Long: cmdutil.LongDesc(`
			Cleans kickoff's local cache of remote skeleton repositories.`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.WithField("cache.dir", kickoff.LocalCacheDir).Info("cleaning cache")

			err := os.RemoveAll(kickoff.LocalCacheDir)
			if err != nil {
				return fmt.Errorf("failed to clean cache: %w", err)
			}

			fmt.Fprintln(streams.Out, color.GreenString("✓"), "Cache cleaned")

			return nil
		},
	}

	return cmd
}
