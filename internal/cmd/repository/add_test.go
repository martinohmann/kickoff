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

func TestAddCmd(t *testing.T) {
	configFile := testutil.NewConfigFileBuilder(t).
		WithRepository("default", "../../testdata/repos/repo1").
		Create()
	defer os.Remove(configFile.Name())

	streams, _, _, _ := cli.NewTestIOStreams()

	f := cmdutil.NewFactoryWithConfigPath(streams, configFile.Name())

	t.Run("repo already exists", func(t *testing.T) {
		cmd := NewAddCmd(f)
		cmd.SetArgs([]string{"default", "../../testdata/repos/repo2"})
		cmd.SetOut(ioutil.Discard)

		err := cmd.Execute()
		require.EqualError(t, err, `repository "default" already exists`)

		config, err := kickoff.LoadConfig(configFile.Name())
		require.NoError(t, err)
		assert.Equal(t, "../../testdata/repos/repo1", config.Repositories["default"])
	})

	t.Run("invalid repository url", func(t *testing.T) {
		cmd := NewAddCmd(f)
		cmd.SetArgs([]string{"new-repo", "invalid\\:"})
		cmd.SetOut(ioutil.Discard)

		err := cmd.Execute()
		require.EqualError(t, err, `invalid repo URL "invalid\\:": parse "invalid\\:": first path segment in URL cannot contain colon`)

		config, err := kickoff.LoadConfig(configFile.Name())
		require.NoError(t, err)
		assert.Equal(t, "../../testdata/repos/repo1", config.Repositories["default"])
	})

	t.Run("add new repo", func(t *testing.T) {
		cmd := NewAddCmd(f)
		cmd.SetArgs([]string{"new-repo", "../../testdata/repos/repo2"})
		cmd.SetOut(ioutil.Discard)

		require.NoError(t, cmd.Execute())

		config, err := kickoff.LoadConfig(configFile.Name())
		require.NoError(t, err)
		assert.Len(t, config.Repositories, 2)
		assert.Equal(t, "../../testdata/repos/repo2", config.Repositories["new-repo"])
	})
}
