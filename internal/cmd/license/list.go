package license

import (
	"fmt"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/license"
	"github.com/spf13/cobra"
)

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
			ctx, cancel := timeoutFlag.Context()
			defer cancel()

			licenses, err := license.List(ctx)
			if err != nil {
				return fmt.Errorf("failed to fetch licenses due to: %v", err)
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
