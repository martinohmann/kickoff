package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/config"
	"github.com/martinohmann/kickoff/internal/file"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	defaultEditor = "vi"
	defaultShell  = "sh"

	editorEnvs = []string{"KICKOFF_EDITOR", "EDITOR"}
)

// NewEditCmd creates a new command that opens the kickoff config in a
// configurable editor so that the user can edit it.
func NewEditCmd(streams cli.IOStreams) *cobra.Command {
	o := &EditOptions{
		IOStreams: streams,
	}

	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit the kickoff config",
		Long: cmdutil.LongDesc(`
			Edit the kickoff config with the editor in the configured the $KICKOFF_EDITOR or $EDITOR environment variable.`),
		Example: cmdutil.Examples(`
			# Edit the default config file
			kickoff config edit

			# Edit custom config file
			kickoff config edit --config custom-config.yaml`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(); err != nil {
				return err
			}

			return o.Run()
		},
	}

	cmdutil.AddConfigFlag(cmd, &o.ConfigPath)

	return cmd
}

// EditOptions holds the options for the edit command.
type EditOptions struct {
	cli.IOStreams
	cmdutil.ConfigFlags
}

// Complete completes the command options.
func (o *EditOptions) Complete() (err error) {
	err = o.ConfigFlags.Complete()
	if err != nil {
		return err
	}

	if o.ConfigPath == "" {
		o.ConfigPath = config.DefaultConfigPath
	}

	o.ConfigPath, err = filepath.Abs(o.ConfigPath)

	return err
}

// Run loads the config file using the configured editor. The config file is
// saved after the editor is closed.
func (o *EditOptions) Run() (err error) {
	var contents []byte

	if !file.Exists(o.ConfigPath) {
		if o.ConfigPath != config.DefaultConfigPath {
			return fmt.Errorf("file %q does not exist", o.ConfigPath)
		}

		contents, err = yaml.Marshal(o.Config)
	} else {
		contents, err = ioutil.ReadFile(o.ConfigPath)
	}

	if err != nil {
		return err
	}

	tmpf, err := ioutil.TempFile("", "kickoff-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}

	tmpfilePath := tmpf.Name()

	defer os.Remove(tmpfilePath)

	log.WithFields(log.Fields{
		"tmpfile":    tmpfilePath,
		"configfile": o.ConfigPath,
	}).Debug("writing config to temporary file")

	err = ioutil.WriteFile(tmpfilePath, contents, 0644)
	if err != nil {
		return err
	}

	err = launchEditor(tmpfilePath)
	if err != nil {
		return err
	}

	// Sanity check: if we fail to load the config from the tmpfile, we
	// consider it invalid and abort without copying it back.
	cfg, err := config.Load(tmpfilePath)
	if err != nil {
		return fmt.Errorf("not saving invalid kickoff config: %w", err)
	}

	log.WithField("config", o.ConfigPath).Info("writing config")

	err = config.Save(&cfg, o.ConfigPath)
	if err != nil {
		return fmt.Errorf("error while saving config file: %w", err)
	}

	fmt.Fprintln(o.Out, "Config saved")

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
