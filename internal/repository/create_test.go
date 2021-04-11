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
		ref, err := Create(t.TempDir() + "/repo")
		require.NoError(t, err)
		require.DirExists(t, ref.SkeletonsPath())
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

func TestCreateWithSkeleton(t *testing.T) {
	t.Parallel()

	t.Run("invalid local path", func(t *testing.T) {
		err := CreateWithSkeleton("invalid\\:", "default")
		require.Error(t, err)
	})

	t.Run("create local repo with skeleton", func(t *testing.T) {
		dir := t.TempDir() + "/repo"
		err := CreateWithSkeleton(dir, "default")
		require.NoError(t, err)
		require.FileExists(t, filepath.Join(dir, "skeletons", "default", "README.md.skel"))
		require.FileExists(t, filepath.Join(dir, "skeletons", "default", kickoff.SkeletonConfigFileName))
	})
}

func TestCreateSkeleton(t *testing.T) {
	t.Parallel()

	t.Run("create skeleton in existing local repository", func(t *testing.T) {
		ref, err := Create(t.TempDir() + "/repo")
		require.NoError(t, err)
		require.NoError(t, CreateSkeleton(ref, "myskeleton"))

		skeletonPath := ref.SkeletonPath("myskeleton")
		require.DirExists(t, skeletonPath)
		require.FileExists(t, filepath.Join(skeletonPath, "README.md.skel"))
		require.FileExists(t, filepath.Join(skeletonPath, kickoff.SkeletonConfigFileName))
	})

	t.Run("cannot create skeleton in nonexistent local repo", func(t *testing.T) {
		ref, err := Create(t.TempDir() + "/repo")
		require.NoError(t, err)
		require.NoError(t, CreateSkeleton(ref, "myskeleton"))
		require.Error(t, CreateSkeleton(ref, "myskeleton"))
	})

	t.Run("cannot overwrite existing skeleton", func(t *testing.T) {
		err := CreateSkeleton(&kickoff.RepoRef{Path: "/non/ex/ist/ent/path"}, "default")
		require.Error(t, err)
	})

	t.Run("cannot create skeleton in remote repo", func(t *testing.T) {
		err := CreateSkeleton(&kickoff.RepoRef{URL: "https://remote.repo"}, "default")
		require.Error(t, err)
	})
}
