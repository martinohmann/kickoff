package skeleton

import (
	"os"
	"testing"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShowFileCmd_Execute_NonexistentRepository(t *testing.T) {
	t.Run("nonexistent repository", func(t *testing.T) {
		configFile := testutil.NewConfigFileBuilder(t).
			WithRepository("default", "../../testdata/repos/repo1").
			Create()
		defer os.Remove(configFile.Name())

		streams, _, _, _ := cli.NewTestIOStreams()
		cmd := NewShowFileCmd(streams)
		cmd.SetArgs([]string{
			"myskeleton",
			"asdf",
			"--config", configFile.Name(),
			"--repository", "nonexistent",
		})

		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("show file", func(t *testing.T) {
		configFile := testutil.NewConfigFileBuilder(t).
			WithRepository("default", "../../testdata/repos/repo1").
			Create()
		defer os.Remove(configFile.Name())

		streams, _, out, _ := cli.NewTestIOStreams()
		cmd := NewShowFileCmd(streams)
		cmd.SetArgs([]string{
			"advanced",
			"README.md.skel",
			"--config", configFile.Name(),
		})

		err := cmd.Execute()
		require.NoError(t, err)

		assert.Contains(t, out.String(), `{{.Project.Name}}`)
	})

	t.Run("nonexistent file", func(t *testing.T) {
		configFile := testutil.NewConfigFileBuilder(t).
			WithRepository("default", "../../testdata/repos/repo1").
			Create()
		defer os.Remove(configFile.Name())

		streams, _, _, _ := cli.NewTestIOStreams()
		cmd := NewShowFileCmd(streams)
		cmd.SetArgs([]string{
			"advanced",
			"nonexistent-file",
			"--config", configFile.Name(),
		})

		err := cmd.Execute()
		require.EqualError(t, err, os.ErrNotExist.Error())
	})

	t.Run("directory", func(t *testing.T) {
		configFile := testutil.NewConfigFileBuilder(t).
			WithRepository("default", "../../testdata/repos/repo1").
			Create()
		defer os.Remove(configFile.Name())

		streams, _, _, _ := cli.NewTestIOStreams()
		cmd := NewShowFileCmd(streams)
		cmd.SetArgs([]string{
			"advanced",
			"{{.Values.filename}}",
			"--config", configFile.Name(),
		})

		err := cmd.Execute()
		require.EqualError(t, err, `"{{.Values.filename}}" is a directory`)
	})
}
