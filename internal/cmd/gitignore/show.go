package gitignore

import (
	"context"
	"strings"

	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/gitignore"
	"github.com/spf13/cobra"
)

// NewShowCmd creates a command that shows the content of one or multiple
// gitignore templates.
func NewShowCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <name> [<name>...]",
		Short: "Fetch a gitignore template",
		Long: cmdutil.LongDesc(`
			Fetches a gitignore template via the GitHub Gitignores API (https://docs.github.com/en/rest/reference/gitignore#get-a-gitignore-template).`),
		Example: cmdutil.Examples(`
			# Fetch a single template
			kickoff gitignore show go

			# Fetch multiple concatenated templates
			kickoff gitignore show go helm hugo`),
		Args: cobra.MinimumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return cmdutil.GitignoreNames(f), cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client := gitignore.NewClient(f.HTTPClient())

			query := strings.Join(args, ",")

			template, err := client.GetTemplate(context.Background(), query)
			if err != nil {
				return err
			}

			_, err = f.IOStreams.Out.Write(template.Content)
			return err
		},
	}

	return cmd
}
