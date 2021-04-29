package repository

import (
	"io"
	"path/filepath"
	"testing"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestCreateCmd(t *testing.T) {
	configPath := testutil.NewConfigFileBuilder(t).
		WithRepository("default", "../../testdata/repos/repo1").
		Create()

	streams, _, _, _ := cli.NewTestIOStreams()

	f := cmdutil.NewFactoryWithConfigPath(streams, configPath)

	t.Run("repo already exists", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "repo")

		cmd := NewCreateCmd(f)
		cmd.SetArgs([]string{"default", dir})
		cmd.SetOut(io.Discard)

		require.Error(t, cmd.Execute())
		require.NoDirExists(t, dir)
	})

	t.Run("creates new repo", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "repo")

		cmd := NewCreateCmd(f)
		cmd.SetArgs([]string{"new-repo", dir})
		cmd.SetOut(io.Discard)

		require.NoError(t, cmd.Execute())
		require.DirExists(t, dir)
	})

	t.Run("creating remote skeletons is not supported", func(t *testing.T) {
		cmd := NewCreateCmd(f)
		cmd.SetArgs([]string{"remote-repo", "https://foo.bar.baz/owner/repo"})
		cmd.SetOut(io.Discard)

		require.Error(t, cmd.Execute())
	})
}
