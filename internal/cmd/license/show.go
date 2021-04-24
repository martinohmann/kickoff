package license

import (
	"context"
	"fmt"

	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/license"
	"github.com/spf13/cobra"
)

// NewShowCmd creates a command that shows the license text of a specific
// license.
func NewShowCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <key>",
		Short: "Fetch a license text",
		Long: cmdutil.LongDesc(`
			Fetches a license text via the GitHub Licenses API (https://docs.github.com/en/rest/reference/licenses#get-a-license).`),
		Example: cmdutil.Examples(`
			# Show MIT license text
			kickoff license show mit`),
		Args: cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return cmdutil.LicenseNames(f), cobra.ShellCompDirectiveDefault
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client := license.NewClient(f.HTTPClient())

			license, err := client.GetLicense(context.Background(), args[0])
			if err != nil {
				return err
			}

			fmt.Fprintln(f.IOStreams.Out, license.Body)

			return nil
		},
	}

	return cmd
}
