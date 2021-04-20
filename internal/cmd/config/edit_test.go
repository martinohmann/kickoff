package config

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/template"
	"github.com/martinohmann/kickoff/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetEditCmdArgs(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		editor   string
		shell    string
		expected []string
	}{
		{
			name:     "default editor and shell",
			path:     "/tmp/foo.yaml",
			expected: []string{"sh", "-c", "vi /tmp/foo.yaml"},
		},
		{
			name:     "editor from env",
			editor:   "nvim",
			path:     "/tmp/foo.yaml",
			expected: []string{"sh", "-c", "nvim /tmp/foo.yaml"},
		},
		{
			name:     "shell from env",
			shell:    "/bin/zsh",
			path:     "/tmp/foo.yaml",
			expected: []string{"/bin/zsh", "-c", "vi /tmp/foo.yaml"},
		},
		{
			name:     "editor and shell from env",
			editor:   "nano",
			shell:    "/bin/ash",
			path:     "/tmp/foo.yaml",
			expected: []string{"/bin/ash", "-c", "nano /tmp/foo.yaml"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			oldEditor := os.Getenv("EDITOR")
			oldShell := os.Getenv("SHELL")

			defer func() {
				os.Setenv("EDITOR", oldEditor)
				os.Setenv("SHELL", oldShell)
			}()

			os.Setenv("EDITOR", test.editor)
			os.Setenv("SHELL", test.shell)

			actual := getEditCmdArgs(test.path)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestEditCmd(t *testing.T) {
	t.Run("invalid editor", func(t *testing.T) {
		oldEditor := os.Getenv("EDITOR")
		oldShell := os.Getenv("SHELL")

		defer func() {
			os.Setenv("EDITOR", oldEditor)
			os.Setenv("SHELL", oldShell)
		}()

		os.Setenv("EDITOR", "./nonexistent")
		os.Setenv("SHELL", "sh")

		configPath := testutil.NewConfigFileBuilder(t).
			WithProjectOwner("johndoe").
			WithRepository("local", "/some/local/path").
			WithRepository("remove", "https://git.john.doe/johndoe/remote-repo").
			WithValues(template.Values{"foo": "bar"}).
			Create()

		configBuf, err := ioutil.ReadFile(configPath)
		require.NoError(t, err)

		streams, _, _, _ := cli.NewTestIOStreams()

		f := cmdutil.NewFactoryWithConfigPath(streams, configPath)

		cmd := NewEditCmd(f)
		cmd.SetOut(ioutil.Discard)

		expectedErrPattern := `error while launching editor command "sh -c ./nonexistent /tmp/kickoff-[0-9]+.yaml": exit status 127`

		err = cmd.Execute()
		require.Error(t, err)

		assert.Regexp(t, expectedErrPattern, err)

		configBuf2, err := ioutil.ReadFile(configPath)
		require.NoError(t, err)

		assert.Equal(t, configBuf, configBuf2, "config file was changed although it should not")
	})
}
