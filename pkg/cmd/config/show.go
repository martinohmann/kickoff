package config

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/martinohmann/kickoff/pkg/cli"
	"github.com/martinohmann/kickoff/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewShowCmd(streams cli.IOStreams) *cobra.Command {
	o := &ShowOptions{IOStreams: streams}

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show the kickoff config",
		Long:  "Show the kickoff config",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(""); err != nil {
				return err
			}

			if err := o.Validate(); err != nil {
				return err
			}

			return o.Run()
		},
	}

	o.ConfigFlags.AddFlags(cmd)
	o.OutputFlags.AddFlags(cmd)

	return cmd
}

type ShowOptions struct {
	cli.IOStreams
	cmdutil.ConfigFlags
	cmdutil.OutputFlags

	Output string
}

func (o *ShowOptions) Run() (err error) {
	var buf []byte

	switch o.Output {
	case "json":
		buf, err = json.MarshalIndent(o.Config, "", "  ")
	default:
		buf, err = yaml.Marshal(o.Config)
	}

	if err != nil {
		return err
	}

	fmt.Fprintln(o.Out, strings.TrimSpace(string(buf)))

	return nil
}
