package cmd

import (
	"errors"
	"fmt"

	"github.com/martinohmann/kickoff/pkg/cli"
	"github.com/martinohmann/kickoff/pkg/cmdutil"
	"github.com/martinohmann/kickoff/pkg/version"
	"github.com/spf13/cobra"
)

var (
	// ErrIllegalVersionFlagCombination is returned if mutual exclusive version
	// format flags are set.
	ErrIllegalVersionFlagCombination = errors.New("--short and --output can't be used together")
)

func NewVersionCmd(streams cli.IOStreams) *cobra.Command {
	o := &VersionOptions{IOStreams: streams}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Displays the version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate()
			if err != nil {
				return err
			}

			return o.Run()
		},
	}

	cmd.Flags().BoolVar(&o.Short, "short", false, "Display short version")

	o.OutputFlags.AddFlags(cmd)

	return cmd
}

type VersionOptions struct {
	cli.IOStreams
	cmdutil.OutputFlags

	Short bool
}

func (o *VersionOptions) Validate() error {
	if o.Short && o.Output != "" {
		return ErrIllegalVersionFlagCombination
	}

	return o.OutputFlags.Validate()
}

func (o *VersionOptions) Run() error {
	v := version.Get()

	if o.Short {
		fmt.Fprintln(o.Out, v.GitVersion)
		return nil
	}

	switch o.Output {
	case "json":
		return cmdutil.RenderJSON(o.Out, v)
	case "yaml":
		return cmdutil.RenderYAML(o.Out, v)
	default:
		fmt.Fprintf(o.Out, "%#v\n", v)
		return nil
	}
}
