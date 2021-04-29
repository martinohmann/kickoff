package skeleton

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShowCmd(t *testing.T) {
	configPath := testutil.NewConfigFileBuilder(t).
		WithRepository("default", "../../testdata/repos/repo1").
		Create()

	streams, _, out, _ := cli.NewTestIOStreams()
	f := cmdutil.NewFactoryWithConfigPath(streams, configPath)

	t.Run("nonexistent repository", func(t *testing.T) {
		cmd := NewShowCmd(f)
		cmd.SetArgs([]string{"myskeleton", "-r", "nonexistent"})
		cmd.SetOut(io.Discard)

		assert.Error(t, cmd.Execute())
	})

	t.Run("nonexistent skeleton", func(t *testing.T) {
		cmd := NewShowCmd(f)
		cmd.SetArgs([]string{"nonexistent"})
		cmd.SetOut(io.Discard)

		require.Error(t, cmd.Execute())
	})

	t.Run("show full", func(t *testing.T) {
		out.Reset()

		cmd := NewShowCmd(f)
		cmd.SetArgs([]string{"minimal"})
		cmd.SetOut(io.Discard)

		require.NoError(t, cmd.Execute())

		output := out.String()

		assert.Regexp(t, `Repository\s+default`, output)
		assert.Regexp(t, `Name\s+minimal`, output)
		assert.Regexp(t, `Description`, output)
		assert.Regexp(t, `Files\s+Values`, output)
	})

	t.Run("show yaml", func(t *testing.T) {
		out.Reset()

		cmd := NewShowCmd(f)
		cmd.SetArgs([]string{"minimal", "-o", "yaml"})
		cmd.SetOut(io.Discard)

		require.NoError(t, cmd.Execute())

		assert.Contains(t, out.String(), "values:\n  foo: bar")
	})

	t.Run("show json", func(t *testing.T) {
		out.Reset()

		cmd := NewShowCmd(f)
		cmd.SetArgs([]string{"minimal", "-o", "json"})
		cmd.SetOut(io.Discard)

		require.NoError(t, cmd.Execute())

		buf := out.Bytes()

		assert.NotEmpty(t, buf)
		var m map[string]interface{}
		require.NoError(t, json.Unmarshal(buf, &m))
		assert.Equal(t, map[string]interface{}{"foo": "bar"}, m["values"])
	})

	t.Run("show file", func(t *testing.T) {
		out.Reset()

		cmd := NewShowCmd(f)
		cmd.SetArgs([]string{"advanced", "README.md.skel"})
		cmd.SetOut(io.Discard)

		require.NoError(t, cmd.Execute())

		assert.Contains(t, out.String(), `{{.Project.Name}}`)
	})

	t.Run("show file as json", func(t *testing.T) {
		out.Reset()

		cmd := NewShowCmd(f)
		cmd.SetArgs([]string{"advanced", "README.md.skel", "-o", "json"})
		cmd.SetOut(io.Discard)

		require.NoError(t, cmd.Execute())

		buf := out.Bytes()

		assert.NotEmpty(t, buf)
		var m map[string]interface{}
		require.NoError(t, json.Unmarshal(buf, &m))
		assert.Equal(t, "README.md.skel", m["relPath"])
	})

	t.Run("show file as yaml", func(t *testing.T) {
		out.Reset()

		cmd := NewShowCmd(f)
		cmd.SetArgs([]string{"advanced", "README.md.skel", "-o", "yaml"})
		cmd.SetOut(io.Discard)

		require.NoError(t, cmd.Execute())

		buf := out.Bytes()

		assert.NotEmpty(t, buf)
		var m map[string]interface{}
		require.NoError(t, yaml.Unmarshal(buf, &m))
		assert.Equal(t, "README.md.skel", m["relPath"])
	})

	t.Run("nonexistent file", func(t *testing.T) {
		cmd := NewShowCmd(f)
		cmd.SetArgs([]string{"advanced", "nonexistent-file"})
		cmd.SetOut(io.Discard)

		err := cmd.Execute()
		require.EqualError(t, err, os.ErrNotExist.Error())
	})

	t.Run("directory", func(t *testing.T) {
		out.Reset()

		cmd := NewShowCmd(f)
		cmd.SetArgs([]string{"advanced", "{{.Values.filename}}"})
		cmd.SetOut(io.Discard)

		err := cmd.Execute()
		require.EqualError(t, err, `{{.Values.filename}} is a directory`)
	})
}
