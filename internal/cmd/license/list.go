package license

import (
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/httpcache"
	"github.com/martinohmann/kickoff/internal/license"
	"github.com/spf13/cobra"
)

// NewListCmd creates a command that lists all open source licenses available
// via the GitHub Licenses API.
func NewListCmd(streams cli.IOStreams) *cobra.Command {
	timeoutFlag := cmdutil.NewDefaultTimeoutFlag()

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List available licenses",
		Long: cmdutil.LongDesc(`
			Lists licenses available via the GitHub Licenses API (https://developer.github.com/v3/licenses/#list-all-licenses).`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := license.NewClient(httpcache.NewClient())

			ctx, cancel := timeoutFlag.Context()
			defer cancel()

			licenses, err := client.ListLicenses(ctx)
			if err != nil {
				return err
			}

			tw := cli.NewTableWriter(streams.Out)
			tw.SetHeader("Key", "Name")

			for _, license := range licenses {
				tw.Append(license.Key, license.Name)
			}

			tw.Render()

			return nil
		},
	}

	timeoutFlag.AddFlag(cmd)

	return cmd
}
