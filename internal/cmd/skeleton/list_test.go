package skeleton

import (
	"io"
	"testing"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListCmd(t *testing.T) {
	configFile := testutil.NewConfigFileBuilder(t).
		WithRepository("default", "../../testdata/repos/repo1").
		Create()

	streams, _, out, _ := cli.NewTestIOStreams()

	f := cmdutil.NewFactoryWithConfigPath(streams, configFile)

	t.Run("default output", func(t *testing.T) {
		out.Reset()

		cmd := NewListCmd(f)
		cmd.SetOut(io.Discard)

		require.NoError(t, cmd.Execute())

		assert.Regexp(t, `Repository\s+Name`, out.String())
		assert.Regexp(t, `default\s+minimal`, out.String())
	})

	t.Run("wide output", func(t *testing.T) {
		out.Reset()

		cmd := NewListCmd(f)
		cmd.SetArgs([]string{"-o", "wide"})
		cmd.SetOut(io.Discard)

		require.NoError(t, cmd.Execute())

		assert.Regexp(t, `Repository\s+Name\s+Path`, out.String())
		assert.Regexp(t, `default\s+minimal\s+`, out.String())
	})

	t.Run("only names", func(t *testing.T) {
		out.Reset()

		cmd := NewListCmd(f)
		cmd.SetArgs([]string{"-o", "name"})
		cmd.SetOut(io.Discard)

		require.NoError(t, cmd.Execute())

		assert.Equal(t, out.String(), "default:advanced\ndefault:minimal\n")
	})
}
