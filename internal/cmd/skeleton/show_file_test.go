package skeleton

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShowFileCmd(t *testing.T) {
	configFile := testutil.NewConfigFileBuilder(t).
		WithRepository("default", "../../testdata/repos/repo1").
		Create()
	defer os.Remove(configFile.Name())

	streams, _, out, _ := cli.NewTestIOStreams()
	f := cmdutil.NewFactoryWithConfigPath(streams, configFile.Name())

	t.Run("nonexistent repository", func(t *testing.T) {
		cmd := NewShowFileCmd(f)
		cmd.SetArgs([]string{"myskeleton", "asdf", "-r", "nonexistent"})
		cmd.SetOut(ioutil.Discard)

		require.Error(t, cmd.Execute())
	})

	t.Run("nonexistent skeleton", func(t *testing.T) {
		cmd := NewShowFileCmd(f)
		cmd.SetArgs([]string{"nonexistent", "README.md.skel"})
		cmd.SetOut(ioutil.Discard)

		require.Error(t, cmd.Execute())
	})

	t.Run("show file", func(t *testing.T) {
		out.Reset()

		cmd := NewShowFileCmd(f)
		cmd.SetArgs([]string{"advanced", "README.md.skel"})
		cmd.SetOut(ioutil.Discard)

		require.NoError(t, cmd.Execute())

		assert.Contains(t, out.String(), `{{.Project.Name}}`)
	})

	t.Run("nonexistent file", func(t *testing.T) {
		cmd := NewShowFileCmd(f)
		cmd.SetArgs([]string{"advanced", "nonexistent-file"})
		cmd.SetOut(ioutil.Discard)

		err := cmd.Execute()
		require.EqualError(t, err, os.ErrNotExist.Error())
	})

	t.Run("directory", func(t *testing.T) {
		out.Reset()

		cmd := NewShowFileCmd(f)
		cmd.SetArgs([]string{"advanced", "{{.Values.filename}}"})
		cmd.SetOut(ioutil.Discard)

		err := cmd.Execute()
		require.EqualError(t, err, `"{{.Values.filename}}" is a directory`)
	})
}
