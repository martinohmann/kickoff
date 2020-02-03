package license

import (
	"fmt"

	"github.com/martinohmann/kickoff/pkg/cli"
	"github.com/martinohmann/kickoff/pkg/cmdutil"
	"github.com/martinohmann/kickoff/pkg/license"
	"github.com/spf13/cobra"
)

func NewShowCmd(streams cli.IOStreams) *cobra.Command {
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
			license, err := license.Get(args[0])
			if err != nil {
				return fmt.Errorf("failed to fetch license text due to: %v", err)
			}

			fmt.Fprintln(streams.Out, license.Body)

			return nil
		},
	}

	return cmd
}
