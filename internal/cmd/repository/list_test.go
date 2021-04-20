package repository

import (
	"io/ioutil"
	"testing"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListCmd(t *testing.T) {
	configPath := testutil.NewConfigFileBuilder(t).
		WithRepository("default", "../../testdata/repos/repo1").
		WithRepository("other", "https://foo.bar.baz/owner/repo").
		Create()

	streams, _, out, _ := cli.NewTestIOStreams()

	f := cmdutil.NewFactoryWithConfigPath(streams, configPath)

	t.Run("default output", func(t *testing.T) {
		out.Reset()

		cmd := NewListCmd(f)
		cmd.SetOut(ioutil.Discard)

		require.NoError(t, cmd.Execute())

		assert.Regexp(t, `Name\s+Type\s+URL\s+Revision`, out.String())
		assert.Regexp(t, `default\s+local`, out.String())
		assert.Regexp(t, `other\s+remote`, out.String())
	})

	t.Run("wide output", func(t *testing.T) {
		out.Reset()

		cmd := NewListCmd(f)
		cmd.SetArgs([]string{"-o", "wide"})
		cmd.SetOut(ioutil.Discard)

		require.NoError(t, cmd.Execute())

		assert.Regexp(t, `Name\s+Type\s+URL\s+Revision\s+Local Path`, out.String())
		assert.Regexp(t, `default\s+local\s+`, out.String())
		assert.Regexp(t, `other\s+remote\s+`, out.String())
	})

	t.Run("only names", func(t *testing.T) {
		out.Reset()

		cmd := NewListCmd(f)
		cmd.SetArgs([]string{"-o", "name"})
		cmd.SetOut(ioutil.Discard)

		require.NoError(t, cmd.Execute())

		assert.Equal(t, out.String(), "default\nother\n")
	})
}
