package config

import (
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/spf13/cobra"
)

// NewShowCmd creates a new command that prints the kickoff config in a
// configurable output format.
func NewShowCmd(f *cmdutil.Factory) *cobra.Command {
	o := &ShowOptions{
		IOStreams: f.IOStreams,
		Config:    f.Config,
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
			kickoff config show --output json`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.Run()
		},
	}

	cmdutil.AddOutputFlag(cmd, &o.Output, "yaml", "json")

	return cmd
}

// ShowOptions holds the options for the show command.
type ShowOptions struct {
	cli.IOStreams

	Config func() (*kickoff.Config, error)

	Output string
}

// Run prints the kickoff config in the configured format.
func (o *ShowOptions) Run() (err error) {
	config, err := o.Config()
	if err != nil {
		return err
	}

	switch o.Output {
	case "json":
		return cmdutil.RenderJSON(o.Out, config)
	default:
		return cmdutil.RenderYAML(o.Out, config)
	}
}
