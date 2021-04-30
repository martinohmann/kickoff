package config

import (
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/creack/pty"
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/template"
	"github.com/martinohmann/kickoff/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEditCmd(t *testing.T) {
	t.Run("invalid editor", func(t *testing.T) {
		defer testutil.Setenv("EDITOR", "./nonexistent")()

		configPath := testutil.NewConfigFileBuilder(t).
			WithProjectOwner("johndoe").
			WithRepository("local", "/some/local/path").
			WithRepository("remote", "https://git.john.doe/johndoe/remote-repo").
			WithValues(template.Values{"foo": "bar"}).
			Create()

		configBuf, err := os.ReadFile(configPath)
		require.NoError(t, err)

		streams, _, _, _ := cli.NewTestIOStreams()

		f := cmdutil.NewFactoryWithConfigPath(streams, configPath)

		cmd := NewEditCmd(f)
		cmd.SetOut(io.Discard)

		err = cmd.Execute()
		require.Error(t, err)

		assert.EqualError(t, err, "editor launch error: fork/exec ./nonexistent: no such file or directory")

		configBuf2, err := os.ReadFile(configPath)
		require.NoError(t, err)

		assert.Equal(t, configBuf, configBuf2, "config file was changed although it should not")
	})

	t.Run("edit and save config", func(t *testing.T) {
		if _, err := exec.LookPath("vi"); err != nil {
			t.Skip("vi not installed, skipping editor test")
		}

		defer testutil.Setenv(kickoff.EnvKeyEditor, "vi")()

		configPath := testutil.NewConfigFileBuilder(t).
			WithProjectOwner("johndoe").
			WithRepository("local", "/some/local/path").
			WithRepository("remote", "https://git.john.doe/johndoe/remote-repo").
			WithValues(template.Values{"foo": "bar"}).
			Create()

		pty, tty, err := pty.Open()
		require.NoError(t, err)
		defer tty.Close()

		streams := cli.IOStreams{
			In:     tty,
			Out:    tty,
			ErrOut: tty,
		}

		f := cmdutil.NewFactoryWithConfigPath(streams, configPath)

		cmd := NewEditCmd(f)
		cmd.SetOut(io.Discard)

		// This goroutine interacts with vi.
		go func() {
			// Delete "local" repository
			pty.Write([]byte("4jdd"))
			// ESC - enter normal mode
			pty.Write([]byte{27})
			// Write and quit - Enter
			pty.Write([]byte(":wq"))
			pty.Write([]byte{13})
		}()

		require.NoError(t, cmd.Execute())

		config, err := kickoff.LoadConfig(configPath)
		require.NoError(t, err)

		expected := &kickoff.Config{
			Project: kickoff.ProjectConfig{
				Host:  "github.com",
				Owner: "johndoe",
			},
			Repositories: map[string]string{
				"remote": "https://git.john.doe/johndoe/remote-repo",
			},
			Values: template.Values{"foo": "bar"},
		}

		assert.Equal(t, expected, config)
	})
}
