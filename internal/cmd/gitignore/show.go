package gitignore

import (
	"context"
	"fmt"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/gitignore"
	"github.com/martinohmann/kickoff/internal/httpcache"
	"github.com/spf13/cobra"
)

// NewShowCmd creates a command that shows the content of one or multiple
// gitignore templates.
func NewShowCmd(streams cli.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Fetch a gitignore template",
		Long: cmdutil.LongDesc(`
			Fetches a gitignore template via the gitignore.io API.

			Check out https://www.gitignore.io for more information about .gitignore templates.`),
		Example: cmdutil.Examples(`
			# Fetch a single template
			kickoff gitignore show go

			# Fetch multiple concatenated templates
			kickoff gitignore show go,helm,hugo`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := gitignore.NewClient(httpcache.NewClient())

			template, err := client.GetTemplate(context.Background(), args[0])
			if err != nil {
				return err
			}

			fmt.Fprintln(streams.Out, string(template.Content))

			return nil
		},
	}

	return cmd
}
