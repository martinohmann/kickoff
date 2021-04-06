package gitignore

import (
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/gitignore"
	"github.com/martinohmann/kickoff/internal/httpcache"
	"github.com/spf13/cobra"
)

// NewListCmd creates a command that lists all gitignore templates available on
// gitignore.io.
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
			client := gitignore.NewClient(httpcache.NewClient())

			ctx, cancel := timeoutFlag.Context()
			defer cancel()

			gitignores, err := client.ListTemplates(ctx)
			if err != nil {
				return err
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
