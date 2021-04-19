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

func TestShowCmd(t *testing.T) {
	configFile := testutil.NewConfigFileBuilder(t).
		WithRepository("default", "../../testdata/repos/repo1").
		Create()
	defer os.Remove(configFile.Name())

	streams, _, out, _ := cli.NewTestIOStreams()
	f := cmdutil.NewFactoryWithConfigPath(streams, configFile.Name())

	t.Run("nonexistent repository", func(t *testing.T) {
		cmd := NewShowCmd(f)
		cmd.SetArgs([]string{"myskeleton", "-r", "nonexistent"})
		cmd.SetOut(ioutil.Discard)

		assert.Error(t, cmd.Execute())
	})

	t.Run("nonexistent skeleton", func(t *testing.T) {
		cmd := NewShowCmd(f)
		cmd.SetArgs([]string{"nonexistent"})
		cmd.SetOut(ioutil.Discard)

		require.Error(t, cmd.Execute())
	})

	t.Run("show full", func(t *testing.T) {
		out.Reset()

		cmd := NewShowCmd(f)
		cmd.SetArgs([]string{"minimal"})
		cmd.SetOut(ioutil.Discard)

		require.NoError(t, cmd.Execute())

		output := out.String()

		assert.Regexp(t, `Name\s+default:minimal`, output)
		assert.Regexp(t, `Values\s+foo: bar`, output)
	})

	t.Run("show yaml", func(t *testing.T) {
		out.Reset()

		cmd := NewShowCmd(f)
		cmd.SetArgs([]string{"minimal", "-o", "yaml"})
		cmd.SetOut(ioutil.Discard)

		require.NoError(t, cmd.Execute())

		assert.Contains(t, out.String(), "values:\n  foo: bar")
	})

	t.Run("show json", func(t *testing.T) {
		out.Reset()

		cmd := NewShowCmd(f)
		cmd.SetArgs([]string{"minimal", "-o", "json"})
		cmd.SetOut(ioutil.Discard)

		require.NoError(t, cmd.Execute())

		assert.Contains(t, out.String(), "  \"values\": {\n    \"foo\": \"bar\"\n  }")
	})
}
