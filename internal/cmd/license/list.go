package license

import (
	"context"
	"fmt"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/license"
	"github.com/spf13/cobra"
)

// NewListCmd creates a command that lists all open source licenses available
// via the GitHub Licenses API.
func NewListCmd(f *cmdutil.Factory) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List available licenses",
		Long: cmdutil.LongDesc(`
			Lists licenses available via the GitHub Licenses API (https://docs.github.com/en/rest/reference/licenses#get-all-commonly-used-licenses).`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := license.NewClient(f.HTTPClient())

			licenses, err := client.ListLicenses(context.Background())
			if err != nil {
				return err
			}

			switch output {
			case "name":
				for _, license := range licenses {
					fmt.Fprintln(f.IOStreams.Out, license.Key)
				}
			default:
				tw := cli.NewTableWriter(f.IOStreams.Out)
				tw.SetHeader("Key", "Name")

				for _, license := range licenses {
					tw.Append(license.Key, license.Name)
				}

				tw.Render()
			}

			return nil
		},
	}

	cmdutil.AddOutputFlag(cmd, &output, "table", "name")

	return cmd
}
