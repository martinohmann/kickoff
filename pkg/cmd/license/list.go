package license

import (
	"github.com/martinohmann/kickoff/pkg/cli"
	"github.com/martinohmann/kickoff/pkg/license"
	"github.com/spf13/cobra"
)

func NewListCmd(streams cli.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List available licenses",
		Long:    "Lists licenses available via the GitHub Licenses API (https://developer.github.com/v3/licenses/#list-all-licenses).",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			licenses, err := license.List()
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

	return cmd
}
