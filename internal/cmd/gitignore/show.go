package gitignore

import (
	"context"
	"fmt"

	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/gitignore"
	"github.com/spf13/cobra"
)

// NewShowCmd creates a command that shows the content of one or multiple
// gitignore templates.
func NewShowCmd(f *cmdutil.Factory) *cobra.Command {
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
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return cmdutil.GitignoreNames(f), cobra.ShellCompDirectiveDefault
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client := gitignore.NewClient(f.HTTPClient())

			template, err := client.GetTemplate(context.Background(), args[0])
			if err != nil {
				return err
			}

			fmt.Fprintln(f.IOStreams.Out, string(template.Content))

			return nil
		},
	}

	return cmd
}
