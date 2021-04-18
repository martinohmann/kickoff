package skeleton

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/repository"
	"github.com/martinohmann/kickoff/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCmd(t *testing.T) {
	tmpdir := t.TempDir() + "/repo"

	repo, err := repository.Create(tmpdir)
	require.NoError(t, err)

	_, err = repo.CreateSkeleton("default")
	require.NoError(t, err)

	myskelDir := filepath.Join(tmpdir, kickoff.SkeletonsDir, "myskel")

	configFile := testutil.NewConfigFileBuilder(t).
		WithRepository("default", tmpdir).
		WithRepository("remote", "https://github.com/martinohmann/kickoff-skeletons").
		Create()
	defer os.Remove(configFile.Name())

	streams, _, _, _ := cli.NewTestIOStreams()

	f := cmdutil.NewFactoryWithConfigPath(streams, configFile.Name())

	t.Run("repository does not exist", func(t *testing.T) {
		cmd := NewCreateCmd(f)
		cmd.SetArgs([]string{"nonexistent", "default"})
		cmd.SetOut(ioutil.Discard)

		err := cmd.Execute()
		assert.EqualError(t, err, `repository "nonexistent" not configured`)
		assert.NoDirExists(t, myskelDir)
	})

	t.Run("remote repo", func(t *testing.T) {
		cmd := NewCreateCmd(f)
		cmd.SetArgs([]string{"remote", "default"})
		cmd.SetOut(ioutil.Discard)

		err := cmd.Execute()
		assert.EqualError(t, err, `creating skeletons in remote repositories is not supported`)
		assert.NoDirExists(t, myskelDir)
	})

	t.Run("skeleton already exists", func(t *testing.T) {
		cmd := NewCreateCmd(f)
		cmd.SetArgs([]string{"default", "default"})
		cmd.SetOut(ioutil.Discard)

		err := cmd.Execute()
		assert.EqualError(t, err, `skeleton "default" already exists in repository "default"`)
		assert.NoDirExists(t, myskelDir)
	})

	t.Run("skeleton can be created", func(t *testing.T) {
		cmd := NewCreateCmd(f)
		cmd.SetArgs([]string{"default", "myskel"})
		cmd.SetOut(ioutil.Discard)

		err := cmd.Execute()
		require.NoError(t, err)
		assert.DirExists(t, myskelDir)
	})
}
