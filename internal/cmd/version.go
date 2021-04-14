package cmd

import (
	"fmt"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/version"
	"github.com/spf13/cobra"
)

// NewVersionCmd creates a command which can print the kickoff version.
func NewVersionCmd(streams cli.IOStreams) *cobra.Command {
	o := &VersionOptions{
		IOStreams:  streams,
		OutputFlag: cmdutil.NewOutputFlag("json", "yaml", "short"),
	}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Displays the version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Validate(); err != nil {
				return err
			}

			return o.Run()
		},
	}

	o.OutputFlag.AddFlag(cmd)

	return cmd
}

// VersionOptions holds the options for the version command.
type VersionOptions struct {
	cli.IOStreams
	cmdutil.OutputFlag
}

// Run prints the version in the provided output format.
func (o *VersionOptions) Run() error {
	v := version.Get()

	switch o.Output {
	case "json":
		return cmdutil.RenderJSON(o.Out, v)
	case "yaml":
		return cmdutil.RenderYAML(o.Out, v)
	case "short":
		fmt.Fprintln(o.Out, v.GitVersion)
	default:
		fmt.Fprintf(o.Out, "%#v\n", v)
	}

	return nil
}
