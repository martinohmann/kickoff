package repository

import (
	"io"
	"path/filepath"
	"testing"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddCmd(t *testing.T) {
	configPath := testutil.NewConfigFileBuilder(t).
		WithRepository("default", "../../testdata/repos/repo1").
		Create()

	streams, _, _, _ := cli.NewTestIOStreams()

	f := cmdutil.NewFactoryWithConfigPath(streams, configPath)

	t.Run("repo already exists", func(t *testing.T) {
		cmd := NewAddCmd(f)
		cmd.SetArgs([]string{"default", "../../testdata/repos/repo2"})
		cmd.SetOut(io.Discard)

		err := cmd.Execute()
		require.EqualError(t, err, `repository "default" already exists`)

		config, err := kickoff.LoadConfig(configPath)
		require.NoError(t, err)
		assert.Equal(t, "../../testdata/repos/repo1", config.Repositories["default"])
	})

	t.Run("invalid repository url", func(t *testing.T) {
		cmd := NewAddCmd(f)
		cmd.SetArgs([]string{"new-repo", "invalid\\:"})
		cmd.SetOut(io.Discard)

		err := cmd.Execute()
		require.EqualError(t, err, `invalid repo URL "invalid\\:": parse "invalid\\:": first path segment in URL cannot contain colon`)

		config, err := kickoff.LoadConfig(configPath)
		require.NoError(t, err)
		assert.Equal(t, "../../testdata/repos/repo1", config.Repositories["default"])
	})

	t.Run("adds new repo, resolves with abspath", func(t *testing.T) {
		cmd := NewAddCmd(f)
		cmd.SetArgs([]string{"new-repo", "../../testdata/repos/repo2"})
		cmd.SetOut(io.Discard)

		absPath, err := filepath.Abs("../../testdata/repos/repo2")
		require.NoError(t, err)

		require.NoError(t, cmd.Execute())

		config, err := kickoff.LoadConfig(configPath)
		require.NoError(t, err)
		assert.Len(t, config.Repositories, 2)
		assert.Equal(t, absPath, config.Repositories["new-repo"])
	})
}
