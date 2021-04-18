package repository

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoveCmd(t *testing.T) {
	configFile := testutil.NewConfigFileBuilder(t).
		WithRepository("default", "../../testdata/repos/repo1").
		WithRepository("other", "../../testdata/repos/repo3").
		Create()
	defer os.Remove(configFile.Name())

	streams, _, _, _ := cli.NewTestIOStreams()

	f := cmdutil.NewFactoryWithConfigPath(streams, configFile.Name())

	t.Run("repo not exists", func(t *testing.T) {
		cmd := NewRemoveCmd(f)
		cmd.SetArgs([]string{"non-existent"})
		cmd.SetOut(ioutil.Discard)

		err := cmd.Execute()
		require.EqualError(t, err, `repository "non-existent" not configured`)

		config, err := kickoff.LoadConfig(configFile.Name())
		require.NoError(t, err)
		assert.Len(t, config.Repositories, 2)
	})

	t.Run("remove a repo", func(t *testing.T) {
		cmd := NewRemoveCmd(f)
		cmd.SetArgs([]string{"other"})
		cmd.SetOut(ioutil.Discard)

		require.NoError(t, cmd.Execute())

		config, err := kickoff.LoadConfig(configFile.Name())
		require.NoError(t, err)
		assert.Len(t, config.Repositories, 1)
	})
}
