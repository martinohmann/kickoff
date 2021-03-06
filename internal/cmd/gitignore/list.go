package gitignore

import (
	"context"
	"fmt"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/gitignore"
	"github.com/spf13/cobra"
)

// NewListCmd creates a command that lists all gitignore templates available
// via the GitHub Gitignores API.
func NewListCmd(f *cmdutil.Factory) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List available gitignores",
		Long: cmdutil.LongDesc(`
			Lists all gitignore templates available via the GitHub Gitignores API (https://docs.github.com/en/rest/reference/gitignore#get-all-gitignore-templates).`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := gitignore.NewClient(f.HTTPClient())

			gitignores, err := client.ListTemplates(context.Background())
			if err != nil {
				return err
			}

			switch output {
			case "name":
				for _, gitignore := range gitignores {
					fmt.Fprintln(f.IOStreams.Out, gitignore)
				}
			default:
				tw := cli.NewTableWriter(f.IOStreams.Out)
				tw.SetHeader("Name")

				for _, gitignore := range gitignores {
					tw.Append(gitignore)
				}

				tw.Render()
			}

			return nil
		},
	}

	cmdutil.AddOutputFlag(cmd, &output, "table", "name")

	return cmd
}
