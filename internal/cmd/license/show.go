package license

import (
	"fmt"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/httpcache"
	"github.com/martinohmann/kickoff/internal/license"
	"github.com/spf13/cobra"
)

// NewShowCmd creates a command that shows the license text of a specific
// license.
func NewShowCmd(streams cli.IOStreams) *cobra.Command {
	timeoutFlag := cmdutil.NewDefaultTimeoutFlag()

	cmd := &cobra.Command{
		Use:   "show <key>",
		Short: "Fetch a license text",
		Long: cmdutil.LongDesc(`
			Fetches a license text via the GitHub Licenses API (https://developer.github.com/v3/licenses/#get-an-individual-license).`),
		Example: cmdutil.Examples(`
			# Show MIT license text
			kickoff license show mit`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := license.NewClient(httpcache.NewClient())

			ctx, cancel := timeoutFlag.Context()
			defer cancel()

			license, err := client.GetLicense(ctx, args[0])
			if err != nil {
				return err
			}

			fmt.Fprintln(streams.Out, license.Body)

			return nil
		},
	}

	timeoutFlag.AddFlag(cmd)

	return cmd
}
