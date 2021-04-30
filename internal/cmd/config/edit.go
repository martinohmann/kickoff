package config

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/ghodss/yaml"
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/editor"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/spf13/cobra"
)

// NewEditCmd creates a new command that opens the kickoff config in a
// configurable editor so that the user can edit it.
func NewEditCmd(f *cmdutil.Factory) *cobra.Command {
	o := &EditOptions{
		IOStreams:  f.IOStreams,
		Config:     f.Config,
		ConfigPath: f.ConfigPath,
	}

	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit the kickoff config",
		Long: fmt.Sprintf(cmdutil.LongDesc(`
			Edit the kickoff config with the editor in the configured the $%s or $EDITOR environment variable.`),
			kickoff.EnvKeyEditor),
		Example: cmdutil.Examples(`
			# Edit the default config file
			kickoff config edit`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.Run()
		},
	}

	return cmd
}

// EditOptions holds the options for the edit command.
type EditOptions struct {
	cli.IOStreams

	Config func() (*kickoff.Config, error)

	ConfigPath string
}

// Run loads the config file using the configured editor. The config file is
// saved after the editor is closed.
func (o *EditOptions) Run() (err error) {
	config, err := o.Config()
	if err != nil {
		return err
	}

	contents, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	editor := editor.New(o.IOStreams)

	edited, err := editor.Edit(contents, "kickoff-*.yaml")
	if err != nil {
		return err
	}

	var newConfig *kickoff.Config

	if err := yaml.Unmarshal(edited, &newConfig); err != nil {
		return err
	}

	err = kickoff.SaveConfig(o.ConfigPath, newConfig)
	if err != nil {
		return fmt.Errorf("error while saving config file: %w", err)
	}

	fmt.Fprintln(o.Out, color.GreenString("✓"), "Config saved")

	return nil
}
