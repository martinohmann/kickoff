package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/ghodss/yaml"
	"github.com/martinohmann/kickoff/pkg/version"
	"github.com/spf13/cobra"
)

var (
	// ErrIllegalVersionFlagCombination is returned if mutual exclusive version
	// format flags are set.
	ErrIllegalVersionFlagCombination = errors.New("--short and --output can't be used together")

	// ErrInvalidOutputFormat is returned if the output format flag contains an
	// invalid value.
	ErrInvalidOutputFormat = errors.New("--output must be 'yaml' or 'json'")
)

func NewVersionCmd() *cobra.Command {
	o := &VersionOptions{}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Displays the version",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate()
			if err != nil {
				return err
			}
			return o.Run()
		},
	}

	cmd.Flags().BoolVar(&o.Short, "short", false, "Display short version")
	cmd.Flags().StringVar(&o.Output, "output", o.Output, "Output format")

	return cmd
}

type VersionOptions struct {
	Short  bool
	Output string
}

func (o *VersionOptions) Validate() error {
	if o.Short && o.Output != "" {
		return ErrIllegalVersionFlagCombination
	}

	if o.Output != "" && o.Output != "yaml" && o.Output != "json" {
		return ErrInvalidOutputFormat
	}

	return nil
}

func (o *VersionOptions) Run() error {
	v := version.Get()

	if o.Short {
		fmt.Fprintln(os.Stdout, v.GitVersion)
		return nil
	}

	switch o.Output {
	case "json":
		buf, err := json.Marshal(v)
		if err != nil {
			return err
		}

		fmt.Fprintln(os.Stdout, string(buf))
	case "yaml":
		buf, err := yaml.Marshal(v)
		if err != nil {
			return err
		}

		fmt.Fprintln(os.Stdout, string(buf))
	default:
		fmt.Fprintf(os.Stdout, "%#v\n", v)
	}

	return nil
}
