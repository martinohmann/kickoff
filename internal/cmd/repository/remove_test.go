package repository

import (
	"io"
	"testing"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoveCmd(t *testing.T) {
	configPath := testutil.NewConfigFileBuilder(t).
		WithRepository("default", "../../testdata/repos/repo1").
		WithRepository("other", "../../testdata/repos/repo3").
		Create()

	streams, _, _, _ := cli.NewTestIOStreams()

	f := cmdutil.NewFactoryWithConfigPath(streams, configPath)

	t.Run("repo not exists", func(t *testing.T) {
		cmd := NewRemoveCmd(f)
		cmd.SetArgs([]string{"non-existent"})
		cmd.SetOut(io.Discard)

		err := cmd.Execute()
		require.EqualError(t, err, `repository "non-existent" not configured`)

		config, err := kickoff.LoadConfig(configPath)
		require.NoError(t, err)
		assert.Len(t, config.Repositories, 2)
	})

	t.Run("remove a repo", func(t *testing.T) {
		cmd := NewRemoveCmd(f)
		cmd.SetArgs([]string{"other"})
		cmd.SetOut(io.Discard)

		require.NoError(t, cmd.Execute())

		config, err := kickoff.LoadConfig(configPath)
		require.NoError(t, err)
		assert.Len(t, config.Repositories, 1)
	})
}
