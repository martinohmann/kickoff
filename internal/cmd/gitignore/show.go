package gitignore

import (
	"fmt"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/gitignore"
	"github.com/spf13/cobra"
)

// NewShowCmd creates a command that shows the content of one or multiple
// gitignore templates.
func NewShowCmd(streams cli.IOStreams) *cobra.Command {
	timeoutFlag := cmdutil.NewDefaultTimeoutFlag()

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
			ctx, cancel := timeoutFlag.Context()
			defer cancel()

			gitignore, err := gitignore.Get(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to fetch gitignore templates due to: %v", err)
			}

			fmt.Fprintln(streams.Out, gitignore)

			return nil
		},
	}

	timeoutFlag.AddFlag(cmd)

	return cmd
}
