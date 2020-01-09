package config

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/martinohmann/kickoff/pkg/cli"
	"github.com/martinohmann/kickoff/pkg/file"
	"github.com/martinohmann/kickoff/pkg/kickoff"
	"github.com/spf13/cobra"
)

var (
	// ErrInvalidOutputFormat is returned if the output format flag contains an
	// invalid value.
	ErrInvalidOutputFormat = errors.New("--output must be 'yaml' or 'json'")
)

func NewShowCmd(streams cli.IOStreams) *cobra.Command {
	o := &ShowOptions{IOStreams: streams, Output: "yaml", ConfigPath: kickoff.DefaultConfigPath}

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show the kickoff config",
		Long:  "Show the kickoff config",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(); err != nil {
				return err
			}

			if err := o.Validate(); err != nil {
				return err
			}

			return o.Run()
		},
	}

	cmd.Flags().StringVar(&o.ConfigPath, "config", o.ConfigPath, "Path to config file")
	cmd.Flags().StringVar(&o.Output, "output", o.Output, "Output format")

	return cmd
}

type ShowOptions struct {
	cli.IOStreams
	ConfigPath string
	Output     string
}

func (o *ShowOptions) Complete() error {
	if o.ConfigPath == "" {
		o.ConfigPath = kickoff.DefaultConfigPath
	}

	return nil
}

func (o *ShowOptions) Validate() error {
	if o.Output != "yaml" && o.Output != "json" {
		return ErrInvalidOutputFormat
	}

	return nil
}

func (o *ShowOptions) Run() (err error) {
	var config *kickoff.Config

	if !file.Exists(o.ConfigPath) {
		if o.ConfigPath == kickoff.DefaultConfigPath {
			config = &kickoff.Config{}
		} else {
			return fmt.Errorf("file %q does not exist", o.ConfigPath)
		}
	} else {
		config, err = kickoff.LoadConfig(o.ConfigPath)
		if err != nil {
			return err
		}
	}

	config.ApplyDefaults("")

	var buf []byte

	switch o.Output {
	case "json":
		buf, err = json.Marshal(config)
	default:
		buf, err = yaml.Marshal(config)
	}

	if err != nil {
		return err
	}

	fmt.Fprintln(o.Out, string(buf))

	return nil
}
