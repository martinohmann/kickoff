package gitignore

import (
	"fmt"

	"github.com/spf13/cobra"
	"kickoff.run/pkg/cli"
	"kickoff.run/pkg/cmdutil"
	"kickoff.run/pkg/gitignore"
)

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
			gitignore, err := gitignore.Get(args[0])
			if err != nil {
				return fmt.Errorf("failed to fetch gitignore templates due to: %v", err)
			}

			fmt.Fprintln(streams.Out, gitignore)

			return nil
		},
	}

	return cmd
}
