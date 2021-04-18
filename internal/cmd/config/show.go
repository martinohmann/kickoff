package config

import (
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewShowCmd creates a new command that prints the kickoff config in a
// configurable output format.
func NewShowCmd(streams cli.IOStreams) *cobra.Command {
	o := &ShowOptions{
		IOStreams: streams,
	}

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show the kickoff config",
		Long: cmdutil.LongDesc(`
			Show the kickoff config`),
		Example: cmdutil.Examples(`
			# Show the default config
			kickoff config show

			# Show the config using different output
			kickoff config show --output json

			# Show a custom config file
			kickoff config show --config custom-config.yaml`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(); err != nil {
				return err
			}

			return o.Run()
		},
	}

	cmdutil.AddConfigFlag(cmd, &o.ConfigPath)
	cmdutil.AddOutputFlag(cmd, &o.Output, "yaml", "json")

	return cmd
}

// ShowOptions holds the options for the show command.
type ShowOptions struct {
	cli.IOStreams
	cmdutil.ConfigFlags

	Output string
}

// Run prints the kickoff config in the configured format.
func (o *ShowOptions) Run() (err error) {
	switch o.Output {
	case "json":
		return cmdutil.RenderJSON(o.Out, o.Config)
	default:
		return cmdutil.RenderYAML(o.Out, o.Config)
	}
}
