package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/martinohmann/kickoff/pkg/boilerplate"
	"github.com/martinohmann/kickoff/pkg/cmdutil"
	"github.com/martinohmann/kickoff/pkg/config"
	"github.com/martinohmann/kickoff/pkg/file"
	"github.com/spf13/cobra"
)

var (
	defaultEditor = "vi"
	defaultShell  = "sh"
)

func NewEditCmd() *cobra.Command {
	o := &EditOptions{}

	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit the kickoff config",
		Long:  "Edit the kickoff config in the configured $EDITOR",
		Args:  cobra.NoArgs,
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

type EditOptions struct {
	ConfigPath string
}

func (o *EditOptions) Complete() (err error) {
	if o.ConfigPath == "" {
		o.ConfigPath = config.DefaultConfigPath
	}

	o.ConfigPath, err = filepath.Abs(o.ConfigPath)

	return err
}

func (o *EditOptions) Run() (err error) {
	var create bool
	if !file.Exists(o.ConfigPath) {
		if o.ConfigPath == config.DefaultConfigPath {
			create = true
		} else {
			return fmt.Errorf("file %q does not exist", o.ConfigPath)
		}
	}

	tmpf, err := ioutil.TempFile("", "kickoff-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}

	defer func() {
		log.WithField("tmpfile", tmpf.Name()).Debug("removing temporary file")
		os.Remove(tmpf.Name())
	}()

	tmpfilePath := tmpf.Name()

	log.WithField("tmpfile", tmpfilePath).Debug("temporary file created")

	contents := boilerplate.DefaultConfigBytes()
	if !create {
		contents, err = ioutil.ReadFile(o.ConfigPath)
		if err != nil {
			return err
		}
	}

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
	_, err = config.Load(tmpfilePath)
	if err != nil {
		return fmt.Errorf("invalid kickoff config: %v", err)
	}

	log.WithFields(log.Fields{
		"tmpfile":    tmpfilePath,
		"configfile": o.ConfigPath,
	}).Debug("copying back config file")

	err = file.Copy(tmpfilePath, o.ConfigPath)
	if err != nil {
		return fmt.Errorf("error while copying back config file: %v", err)
	}

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
		return fmt.Errorf("error while launching editor command %q: %v", commandLine, err)
	}

	return nil
}

func getEditCmdArgs(path string) []string {
	cmdArgs := []string{detectEditor(), path}

	return append(detectShell(), strings.Join(cmdArgs, " "))
}

func detectEditor() string {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = defaultEditor
	}

	return editor
}

func detectShell() []string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = defaultShell
	}

	return []string{shell, "-c"}
}
