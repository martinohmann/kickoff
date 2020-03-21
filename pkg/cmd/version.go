package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ghodss/yaml"
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
		buf, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return err
		}

		fmt.Fprintln(o.Out, string(buf))
	case "yaml":
		buf, err := yaml.Marshal(v)
		if err != nil {
			return err
		}

		fmt.Fprintln(o.Out, string(buf))
	default:
		fmt.Fprintf(o.Out, "%#v\n", v)
	}

	return nil
}
