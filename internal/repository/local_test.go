package repository

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/martinohmann/kickoff/internal/skeleton"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalRepository_GetSkeleton(t *testing.T) {
	info := skeleton.RepoInfo{Path: "../testdata/repos/repo1"}

	repo, err := NewLocalRepository(info)
	require.NoError(t, err)

	abspath, err := filepath.Abs("../testdata/repos/repo1")
	require.NoError(t, err)

	t.Run("can retrieve info about a single skeleton", func(t *testing.T) {
		info, err := repo.GetSkeleton(context.Background(), "minimal")
		require.NoError(t, err)

		assert.Equal(t, abspath, info.Repo.Path)
		assert.Equal(t, filepath.Join(abspath, "skeletons", "minimal"), info.Path)
		assert.Equal(t, "minimal", info.Name)
	})

	t.Run("returns SkeletonNotFoundError if skeleton does not exist", func(t *testing.T) {
		_, err := repo.GetSkeleton(context.Background(), "nonexistent")
		require.Error(t, err)

		assert.IsType(t, SkeletonNotFoundError{}, err)
	})
}

func TestLocalRepository_ListSkeletons(t *testing.T) {
	t.Run("can list all skeletons", func(t *testing.T) {
		info := skeleton.RepoInfo{Path: "../testdata/repos/repo1"}

		repo, err := NewLocalRepository(info)
		require.NoError(t, err)

		infos, err := repo.ListSkeletons(context.Background())
		require.NoError(t, err)

		require.Len(t, infos, 2)
		assert.Equal(t, "advanced", infos[0].Name)
		assert.Equal(t, "minimal", infos[1].Name)
	})

	t.Run("listing nonexistent repository causes error", func(t *testing.T) {
		repo, err := NewLocalRepository(skeleton.RepoInfo{})
		require.NoError(t, err)

		_, err = repo.ListSkeletons(context.Background())
		require.Error(t, err)
	})
}
