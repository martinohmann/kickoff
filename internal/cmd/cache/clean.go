package cache

import (
	"fmt"
	"os"

	"github.com/kirsle/configdir"
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
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
			cacheDir := configdir.LocalCache("kickoff")

			log.WithField("cache.dir", cacheDir).Info("cleaning cache")

			err := os.RemoveAll(cacheDir)
			if err != nil {
				return fmt.Errorf("failed to clean cache: %w", err)
			}

			fmt.Fprintln(streams.Out, "Cache cleaned.")

			return nil
		},
	}

	return cmd
}
