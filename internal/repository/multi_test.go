package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMultiRepository(t *testing.T) {
	t.Run("fails to create if one of the repositories is invalid", func(t *testing.T) {
		_, err := NewMultiRepository(map[string]string{
			"repo1": "../testdata/repos/repo1",
			"repo2": "\nhttpxd::/asdf\\invalid",
		})
		require.Error(t, err)
	})

	t.Run("fails to create if one of the repositories has an empty key", func(t *testing.T) {
		_, err := NewMultiRepository(map[string]string{
			"": "../testdata/repos/repo1",
		})
		require.Error(t, err)
	})

	t.Run("fails to create if repo map is empty", func(t *testing.T) {
		_, err := NewMultiRepository(nil)
		require.Error(t, err)
	})
}

func TestMultiRepository_GetSkeleton(t *testing.T) {
	repo, err := NewMultiRepository(map[string]string{
		"repo1": "../testdata/repos/repo1",
		"repo2": "../testdata/repos/repo2",
	})
	require.NoError(t, err)

	t.Run("retrieves unique skeleton from backing repository", func(t *testing.T) {
		info, err := repo.GetSkeleton(context.Background(), "advanced")
		require.NoError(t, err)

		assert.Equal(t, "advanced", info.Name)
		assert.Equal(t, "repo1", info.Repo.Name)
	})

	t.Run("returns SkeletonNotFoundError if skeleton does not exist", func(t *testing.T) {
		_, err := repo.GetSkeleton(context.Background(), "nonexistent")
		require.Error(t, err)

		assert.IsType(t, SkeletonNotFoundError{}, err)
	})

	t.Run("returns error if skeleton name is ambiguous", func(t *testing.T) {
		_, err := repo.GetSkeleton(context.Background(), "minimal")
		require.Error(t, err)

		assert.Equal(t, `skeleton "minimal" found in multiple repositories: repo1, repo2. explicitly provide <repo-name>:minimal to select one`, err.Error())
	})

	t.Run("retrieves skeleton by repo name", func(t *testing.T) {
		info, err := repo.GetSkeleton(context.Background(), "repo2:minimal")
		require.NoError(t, err)

		assert.Equal(t, "minimal", info.Name)
		assert.Equal(t, "repo2", info.Repo.Name)
	})

	t.Run("returns error if repo name is unknown", func(t *testing.T) {
		_, err := repo.GetSkeleton(context.Background(), "nonexistent-repo:minimal")
		require.Error(t, err)
	})

	t.Run("invalid repository causes error on GetSkeleton", func(t *testing.T) {
		repo, err := NewMultiRepository(map[string]string{
			"invalid": "multi_test.go", // file instead of dir
		})
		require.NoError(t, err)

		_, err = repo.GetSkeleton(context.Background(), "nonexistent")
		require.Error(t, err)

		if _, ok := err.(SkeletonNotFoundError); ok {
			t.Fatal("expected an error different from SkeletonNotFoundError")
		}
	})
}

func TestMultiRepository_ListSkeletons(t *testing.T) {
	repo, err := NewMultiRepository(map[string]string{
		"repo1": "../testdata/repos/repo1",
		"repo2": "../testdata/repos/repo2",
	})
	require.NoError(t, err)

	t.Run("can list skeletons from all repositories", func(t *testing.T) {
		infos, err := repo.ListSkeletons(context.Background())
		require.NoError(t, err)

		require.Len(t, infos, 3)
		assert.Equal(t, "advanced", infos[0].Name)
		assert.Equal(t, "repo1", infos[0].Repo.Name)
		assert.Equal(t, "minimal", infos[1].Name)
		assert.Equal(t, "repo1", infos[1].Repo.Name)
		assert.Equal(t, "minimal", infos[2].Name)
		assert.Equal(t, "repo2", infos[2].Repo.Name)
	})

	t.Run("invalid repository causes error on ListSkeletons", func(t *testing.T) {
		repo, err := NewMultiRepository(map[string]string{
			"invalid": "multi_test.go", // file instead of dir
		})
		require.NoError(t, err)

		_, err = repo.ListSkeletons(context.Background())
		require.Error(t, err)
	})
}
