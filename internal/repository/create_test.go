package repository

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	t.Run("create local repo", func(t *testing.T) {
		dir := t.TempDir() + "/repo"

		_, err := Create(dir)
		require.NoError(t, err)
		require.DirExists(t, filepath.Join(dir, kickoff.SkeletonsDir))
	})

	t.Run("returns error on invalid path", func(t *testing.T) {
		_, err := Create("invalid\\:")
		require.Error(t, err)
	})

	t.Run("cannot create remote repo", func(t *testing.T) {
		_, err := Create("https://remote.repo")
		require.Error(t, err)
	})

	t.Run("cannot create local repo if path exist", func(t *testing.T) {
		dir := t.TempDir() + "/repo"
		require.NoError(t, os.MkdirAll(dir, 0755))

		_, err := Create(dir)
		require.Error(t, err)
	})
}
