package cmd

import (
	"fmt"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewCompletionCmd creates a new command which can set up shell completion.
func NewCompletionCmd(streams cli.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion <shell>",
		Short: "Shell completion",
		Long: cmdutil.LongDesc(`
			Configure your shell to load kickoff completions.
		`),
		Example: cmdutil.Examples(`
			# Add to your ~/.bashrc for bash completion:
			. <(kickoff completion bash)

			# Add to your ~/.zshrc for zsh completion:
			. <(kickoff completion zsh)
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			shell := args[0]
			rootCmd := cmd.Root()

			switch shell {
			case "bash":
				return rootCmd.GenBashCompletion(streams.Out)
			case "zsh":
				err := rootCmd.GenZshCompletion(streams.Out)
				if err != nil {
					return err
				}

				fmt.Fprintln(streams.Out, "compdef _kickoff kickoff")
			default:
				return fmt.Errorf("unsupported shell: %s", shell)
			}
			return nil
		},
	}

	return cmd
}
