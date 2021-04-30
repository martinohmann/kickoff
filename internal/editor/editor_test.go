package editor

import (
	"os/exec"
	"testing"

	"github.com/creack/pty"
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEditor_Edit(t *testing.T) {
	if _, err := exec.LookPath("vi"); err != nil {
		t.Skip("vi not installed, skipping editor test")
	}

	defer testutil.Unsetenv(kickoff.EnvKeyEditor)()
	defer testutil.Unsetenv("EDITOR")()
	defer testutil.Unsetenv("VISUAL")()

	pty, tty, err := pty.Open()
	require.NoError(t, err)
	defer tty.Close()

	streams := cli.IOStreams{
		In:     tty,
		Out:    tty,
		ErrOut: tty,
	}

	editor := New(streams)

	// This goroutine interacts with vi.
	go func() {
		// Enter append mode
		pty.Write([]byte("A"))
		// Append some text
		pty.Write([]byte(" and some changes"))
		// ESC - enter normal mode
		pty.Write([]byte{27})
		// Write and quit - Enter
		pty.Write([]byte(":wq"))
		pty.Write([]byte{13})
	}()

	out, err := editor.Edit([]byte(`input`), "*.yaml")
	require.NoError(t, err)
	assert.Equal(t, "input and some changes\n", string(out))
}

func TestDetect(t *testing.T) {
	t.Run("detects editor from env", func(t *testing.T) {
		defer testutil.Setenv(kickoff.EnvKeyEditor, "fancyeditor")()

		assert.Equal(t, "fancyeditor", Detect())
	})

	t.Run("falls back to vi", func(t *testing.T) {
		defer testutil.Unsetenv(kickoff.EnvKeyEditor)()
		defer testutil.Unsetenv("EDITOR")()
		defer testutil.Unsetenv("VISUAL")()

		assert.Equal(t, "vi", Detect())
	})
}
