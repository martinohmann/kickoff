package gitignore

import (
	"fmt"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/gitignore"
	"github.com/spf13/cobra"
)

func NewListCmd(streams cli.IOStreams) *cobra.Command {
	timeoutFlag := cmdutil.NewDefaultTimeoutFlag()

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List available gitignores",
		Long: cmdutil.LongDesc(`
			Lists gitignores available via the gitignore.io API.

			Check out https://www.gitignore.io for more information about .gitignore templates.`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := timeoutFlag.Context()
			defer cancel()

			gitignores, err := gitignore.List(ctx)
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

	timeoutFlag.AddFlag(cmd)

	return cmd
}
