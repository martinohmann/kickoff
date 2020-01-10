package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/martinohmann/kickoff/pkg/cli"
	"github.com/martinohmann/kickoff/pkg/config"
	"github.com/martinohmann/kickoff/pkg/file"
	"github.com/spf13/cobra"
)

var (
	// ErrInvalidOutputFormat is returned if the output format flag contains an
	// invalid value.
	ErrInvalidOutputFormat = errors.New("--output must be 'yaml' or 'json'")
)

func NewShowCmd(streams cli.IOStreams) *cobra.Command {
	o := &ShowOptions{IOStreams: streams, Output: "yaml", ConfigPath: config.DefaultConfigPath}

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
		o.ConfigPath = config.DefaultConfigPath
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
	var cfg config.Config

	if !file.Exists(o.ConfigPath) {
		if o.ConfigPath != config.DefaultConfigPath {
			return fmt.Errorf("file %q does not exist", o.ConfigPath)
		}
	} else {
		cfg, err = config.Load(o.ConfigPath)
		if err != nil {
			return err
		}
	}

	cfg.ApplyDefaults("")

	var buf []byte

	switch o.Output {
	case "json":
		buf, err = json.MarshalIndent(cfg, "", "  ")
	default:
		buf, err = yaml.Marshal(cfg)
	}

	if err != nil {
		return err
	}

	fmt.Fprintln(o.Out, strings.TrimSpace(string(buf)))

	return nil
}
