package skeleton

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/repository"
	"github.com/martinohmann/kickoff/internal/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCmd(t *testing.T) {
	tmpdir := t.TempDir() + "/repo"

	ref, err := repository.Create(tmpdir)
	require.NoError(t, err)
	require.NoError(t, repository.CreateSkeleton(ref, "default"))

	myskelDir := filepath.Join(tmpdir, kickoff.SkeletonsDir, "myskel")

	configFile := testutil.NewConfigFileBuilder(t).
		WithRepository("default", tmpdir).
		WithRepository("remote", "https://github.com/martinohmann/kickoff-skeletons").
		Create()
	defer os.Remove(configFile.Name())

	t.Run("empty repo name", func(t *testing.T) {
		cmd := newCreateCmd()
		cmd.SetArgs([]string{"", "myskel", "--config", configFile.Name()})

		err := cmd.Execute()
		assert.EqualError(t, err, "repository name must not be empty")
		assert.NoDirExists(t, myskelDir)
	})

	t.Run("empty skeleton name", func(t *testing.T) {
		cmd := newCreateCmd()
		cmd.SetArgs([]string{"myrepo", "", "--config", configFile.Name()})

		err := cmd.Execute()
		assert.EqualError(t, err, "skeleton name must not be empty")
		assert.NoDirExists(t, myskelDir)
	})

	t.Run("repository does not exist", func(t *testing.T) {
		cmd := newCreateCmd()
		cmd.SetArgs([]string{"nonexistent", "default", "--config", configFile.Name()})

		err := cmd.Execute()
		assert.EqualError(t, err, `repository "nonexistent" not configured`)
		assert.NoDirExists(t, myskelDir)
	})

	t.Run("remote repo", func(t *testing.T) {
		cmd := newCreateCmd()
		cmd.SetArgs([]string{"remote", "default", "--config", configFile.Name()})

		err := cmd.Execute()
		assert.EqualError(t, err, `creating skeletons in remote repositories is not supported`)
		assert.NoDirExists(t, myskelDir)
	})

	t.Run("skeleton already exists", func(t *testing.T) {
		cmd := newCreateCmd()
		cmd.SetArgs([]string{"default", "default", "--config", configFile.Name()})

		err := cmd.Execute()
		assert.EqualError(t, err, `skeleton "default" already exists in repository "default"`)
		assert.NoDirExists(t, myskelDir)
	})

	t.Run("skeleton can be created", func(t *testing.T) {
		cmd := newCreateCmd()
		cmd.SetArgs([]string{"default", "myskel", "--config", configFile.Name()})

		err := cmd.Execute()
		require.NoError(t, err)
		assert.DirExists(t, myskelDir)
	})
}

func newCreateCmd() *cobra.Command {
	streams, _, out, errOut := cli.NewTestIOStreams()
	cmd := NewCreateCmd(streams)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	return cmd
}
