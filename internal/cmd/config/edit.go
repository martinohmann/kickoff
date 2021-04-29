package config

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/ghodss/yaml"
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/kickoff"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	defaultEditor = "vi"
	defaultShell  = "sh"

	editorEnvs = []string{kickoff.EnvKeyEditor, "EDITOR"}
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

	tmpf, err := os.CreateTemp("", "kickoff-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}

	tmpfilePath := tmpf.Name()

	defer os.Remove(tmpfilePath)

	log.WithFields(log.Fields{
		"tmpfile":    tmpfilePath,
		"configfile": o.ConfigPath,
	}).Debug("writing config to temporary file")

	if err := os.WriteFile(tmpfilePath, contents, 0644); err != nil {
		return err
	}

	if err := launchEditor(tmpfilePath); err != nil {
		return err
	}

	// Sanity check: if we fail to load the config from the tmpfile, we
	// consider it invalid and abort without copying it back.
	config, err = kickoff.LoadConfig(tmpfilePath)
	if err != nil {
		return fmt.Errorf("not saving invalid kickoff config: %w", err)
	}

	err = kickoff.SaveConfig(o.ConfigPath, config)
	if err != nil {
		return fmt.Errorf("error while saving config file: %w", err)
	}

	fmt.Fprintln(o.Out, color.GreenString("✓"), "Config saved")

	return nil
}

func launchEditor(path string) error {
	args := getEditCmdArgs(path)

	commandLine := strings.Join(args, " ")

	log.WithField("commandLine", commandLine).Debug("launching editor")

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error while launching editor command %q: %w", commandLine, err)
	}

	return nil
}

func getEditCmdArgs(path string) []string {
	cmdArgs := []string{detectEditor(), path}

	return append(detectShell(), strings.Join(cmdArgs, " "))
}

func detectEditor() string {
	for _, env := range editorEnvs {
		if editor := os.Getenv(env); editor != "" {
			return editor
		}
	}

	return defaultEditor
}

func detectShell() []string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = defaultShell
	}

	return []string{shell, "-c"}
}
