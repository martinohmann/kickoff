package gitignore

import (
	"fmt"

	"github.com/martinohmann/kickoff/pkg/cli"
	"github.com/martinohmann/kickoff/pkg/cmdutil"
	"github.com/martinohmann/kickoff/pkg/gitignore"
	"github.com/spf13/cobra"
)

func NewListCmd(streams cli.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List available gitignores",
		Long: cmdutil.LongDesc(`
			Lists gitignores available via the gitignore.io API.

			Check out https://www.gitignore.io for more information about .gitignore templates.`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			gitignores, err := gitignore.List()
			if err != nil {
				return fmt.Errorf("failed to fetch gitignore templates due to: %v", err)
			}

			tw := cli.NewTableWriter(streams.Out)
			tw.SetHeader("Name")

			for _, gitignore := range gitignores {
				tw.Append(gitignore)
			}

			tw.Render()

			return nil
		},
	}

	return cmd
}
