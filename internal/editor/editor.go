package editor

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/google/shlex"
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/kickoff"
	log "github.com/sirupsen/logrus"
)

var (
	bom            = []byte{0xef, 0xbb, 0xbf}
	fallbackEditor = "vi"
	editorEnvs     = []string{kickoff.EnvKeyEditor, "EDITOR", "VISUAL"}
)

// Detect the editor by looking into the KICKOFF_EDITOR, EDITOR and VISUAL
// environment variables. If none of these contains an editor command "vi" is
// used as a fallback.
func Detect() string {
	for _, env := range editorEnvs {
		if editor := os.Getenv(env); editor != "" {
			log.WithFields(log.Fields{
				"env":    env,
				"editor": editor,
			}).Debug("using editor configured via env")

			return editor
		}
	}

	log.WithField("editor", fallbackEditor).Debug("using fallback editor")

	return fallbackEditor
}

// Editor can prompt a user to manipulate text by opening it with an editor.
type Editor struct {
	cli.IOStreams
	Command string
}

// New creates a new *Editor.
func New(streams cli.IOStreams) *Editor {
	return &Editor{IOStreams: streams}
}

// Edit writes contents to a temporary file and opens it with an editor . A
// pattern (e.g. '*.yaml') can be provided to control the extension of the
// temporary file that is created and then opened with the editor. This is
// useful as a hint for the editor so that it can choose a suitable syntax
// highlighting.
// Returns the byte slice with the edited result after the user closed the
// editor.
func (e *Editor) Edit(contents []byte, pattern string) ([]byte, error) {
	f, err := os.CreateTemp("", pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(f.Name())

	if err := writeClose(f, contents); err != nil {
		return nil, fmt.Errorf("failed to write temporary file: %w", err)
	}

	if err := e.launch(f.Name()); err != nil {
		return nil, fmt.Errorf("editor launch error: %w", err)
	}

	return readStrip(f.Name())
}

func (e *Editor) launch(filename string) error {
	editorCommand := e.Command
	if editorCommand == "" {
		editorCommand = Detect()
	}

	args, err := shlex.Split(editorCommand)
	if err != nil {
		return err
	}

	args = append(args, filename)

	log.WithField("command", args).Debug("launching editor")

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = e.In
	cmd.Stdout = e.Out
	cmd.Stderr = e.ErrOut

	return cmd.Run()
}

func writeClose(w io.WriteCloser, contents []byte) error {
	defer w.Close()
	// Write utf8 BOM header.
	//
	// The reason why we do this is because notepad.exe on Windows determines the
	// encoding of an "empty" text file by the locale, for example, GBK in China,
	// while golang string only handles utf8 well. However, a text file with utf8
	// BOM header is not considered "empty" on Windows, and the encoding will then
	// be determined utf8 by notepad.exe, instead of GBK or other encodings.
	//
	// This was copied from the survey source:
	// https://github.com/AlecAivazis/survey/blob/b70520c4e71ed5077b3b285b7ec3ad8dd16e7e78/editor.go#L153-L157
	if _, err := w.Write(bom); err != nil {
		return err
	}

	if _, err := w.Write(contents); err != nil {
		return err
	}

	return nil
}

func readStrip(filename string) ([]byte, error) {
	edited, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read back file: %w", err)
	}

	// Strip BOM header that was added previously in writeClose.
	return bytes.TrimPrefix(edited, bom), nil
}
