package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"

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

            Bash:

              $ source <(kickoff completion bash)

              # To load completions for each session, execute once:
              # Linux:
              $ kickoff completion bash > /etc/bash_completion.d/kickoff
              # macOS:
              $ kickoff completion bash > /usr/local/etc/bash_completion.d/kickoff

            Zsh:

              # If shell completion is not already enabled in your environment,
              # you will need to enable it.  You can execute the following once:

              $ echo "autoload -U compinit; compinit" >> ~/.zshrc

              $ source <(kickoff completion zsh)

              # To load completions for each session, execute once:
              $ kickoff completion zsh > "${fpath[1]}/_kickoff"

              # You will need to start a new shell for this setup to take effect.

            fish:

              $ kickoff completion fish | source

              # To load completions for each session, execute once:
              $ kickoff completion fish > ~/.config/fish/completions/kickoff.fish

            PowerShell:

              PS> kickoff completion powershell | Out-String | Invoke-Expression

              # To load completions for every new session, run:
              PS> kickoff completion powershell > kickoff.ps1
              # and source this file from your PowerShell profile.
        `),
		Example: cmdutil.Examples(`
			# Add to your ~/.bashrc for bash completion:
			source <(kickoff completion bash)

			# Add to your ~/.zshrc for zsh completion:
			source <(kickoff completion zsh)
		`),
		DisableFlagsInUseLine: true,
		Args:                  cobra.ExactArgs(1),
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		RunE: func(cmd *cobra.Command, args []string) error {
			shell := args[0]
			rootCmd := cmd.Root()

			switch shell {
			case "bash":
				// Cobra currently does not handle colons in completions
				// properly. Full skeleton names have the form `repo:name` and
				// the completion of these stops at the colon with the
				// generated completion script. To fix this we patch the
				// relevant lines from the generated script before returning
				// it.
				//
				// See issue comment:
				// https://github.com/spf13/cobra/issues/1355#issuecomment-787222293
				var buf bytes.Buffer

				if err := rootCmd.GenBashCompletion(&buf); err != nil {
					return err
				}

				replacer := strings.NewReplacer(
					`_init_completion -s`,
					`_init_completion -s -n ":"`,
					`__kickoff_init_completion -n "="`,
					`__kickoff_init_completion -n ":="`,
				)

				fmt.Fprintln(streams.Out, replacer.Replace(buf.String()))
			case "zsh":
				if err := rootCmd.GenZshCompletion(streams.Out); err != nil {
					return err
				}

				// Append compdef so that directly sourcing the completion
				// output works.
				fmt.Fprintln(streams.Out, "compdef _kickoff kickoff")
			case "fish":
				return rootCmd.GenFishCompletion(os.Stdout, true)
			case "powershell":
				return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
			default:
				return fmt.Errorf("unsupported shell: %s", shell)
			}
			return nil
		},
	}

	return cmd
}
