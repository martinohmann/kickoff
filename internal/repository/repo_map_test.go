package repository

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenMap(t *testing.T) {
	t.Parallel()

	t.Run("fails to create if one of the repositories is invalid", func(t *testing.T) {
		_, err := OpenMap(context.Background(), map[string]string{
			"repo1": "../testdata/repos/repo1",
			"repo2": "\nhttpxd::/asdf\\invalid",
		}, nil)
		require.Error(t, err)
	})

	t.Run("fails to create if one of the repositories has an empty key", func(t *testing.T) {
		_, err := OpenMap(context.Background(), map[string]string{
			"": "../testdata/repos/repo1",
		}, nil)
		require.Error(t, err)
	})

	t.Run("fails to create if repo map is empty", func(t *testing.T) {
		_, err := OpenMap(context.Background(), nil, nil)
		require.Error(t, err)
	})

	t.Run("invalid local repository", func(t *testing.T) {
		_, err := OpenMap(context.Background(), map[string]string{
			"invalid": "./multi_test.go", // file instead of dir
		}, nil)
		require.Error(t, err)
	})
}

func TestRepositoryMap_GetSkeleton(t *testing.T) {
	t.Parallel()

	repo, err := OpenMap(context.Background(), map[string]string{
		"repo1": "../testdata/repos/repo1",
		"repo2": "../testdata/repos/repo2",
	}, nil)
	require.NoError(t, err)

	t.Run("retrieves unique skeleton from backing repository", func(t *testing.T) {
		ref, err := repo.GetSkeleton("advanced")
		require.NoError(t, err)

		assert.Equal(t, "advanced", ref.Name)
		assert.Equal(t, "repo1", ref.Repo.Name)
	})

	t.Run("returns SkeletonNotFoundError if skeleton does not exist", func(t *testing.T) {
		_, err := repo.GetSkeleton("nonexistent")
		require.EqualError(t, err, `skeleton "nonexistent" not found`)
	})

	t.Run("returns error if skeleton name is ambiguous", func(t *testing.T) {
		_, err := repo.GetSkeleton("minimal")
		require.Error(t, err)

		assert.Equal(t, `skeleton "minimal" found in multiple repositories: repo1, repo2. explicitly provide <repo-name>:minimal to select one`, err.Error())
	})

	t.Run("retrieves skeleton by repo name", func(t *testing.T) {
		ref, err := repo.GetSkeleton("repo2:minimal")
		require.NoError(t, err)

		assert.Equal(t, "minimal", ref.Name)
		assert.Equal(t, "repo2", ref.Repo.Name)
	})

	t.Run("returns error if repo name is unknown", func(t *testing.T) {
		_, err := repo.GetSkeleton("nonexistent-repo:minimal")
		require.Error(t, err)
	})
}

func TestRepositoryMap_ListSkeletons(t *testing.T) {
	t.Parallel()

	t.Run("can list skeletons from all repositories", func(t *testing.T) {
		repo, err := OpenMap(context.Background(), map[string]string{
			"repo1": "../testdata/repos/repo1",
			"repo2": "../testdata/repos/repo2",
		}, nil)
		require.NoError(t, err)

		refs, err := repo.ListSkeletons()
		require.NoError(t, err)

		require.Len(t, refs, 3)
		assert.Equal(t, "advanced", refs[0].Name)
		assert.Equal(t, "repo1", refs[0].Repo.Name)
		assert.Equal(t, "minimal", refs[1].Name)
		assert.Equal(t, "repo1", refs[1].Repo.Name)
		assert.Equal(t, "minimal", refs[2].Name)
		assert.Equal(t, "repo2", refs[2].Repo.Name)
	})
}

func TestRepositoryMap_CreateSkeleton(t *testing.T) {
	t.Parallel()

	t.Run("creates skeleton without repo name if there is only one repo", func(t *testing.T) {
		repo := createTestMultiRepo(t, "repo1")

		ref, err := repo.CreateSkeleton("myskeleton")
		require.NoError(t, err)

		require.DirExists(t, ref.Repo.SkeletonPath("myskeleton"))
	})

	t.Run("returns error if there are multiple repos and skeleton name is ambiguous", func(t *testing.T) {
		repo := createTestMultiRepo(t, "repo1", "repo2")

		_, err := repo.CreateSkeleton("myskeleton")
		require.EqualError(t, err, `ambiguous skeleton name "myskeleton": explicitly provide <repo-name>:myskeleton to select a repository`)
	})

	t.Run("creates skeleton with fully qualified name", func(t *testing.T) {
		repo := createTestMultiRepo(t, "repo1", "repo2")

		ref, err := repo.CreateSkeleton("repo2:myskeleton")
		require.NoError(t, err)
		require.Equal(t, "repo2", ref.Repo.Name)
		require.DirExists(t, ref.Repo.SkeletonPath("myskeleton"))
	})

	t.Run("returns error if named repository does not exist", func(t *testing.T) {
		repo := createTestMultiRepo(t, "repo1")

		_, err := repo.CreateSkeleton("repo2:myskeleton")
		require.EqualError(t, err, `no skeleton repository configured with name "repo2"`)
	})
}

func TestRepositoryMap_LoadSkeleton(t *testing.T) {
	t.Parallel()

	t.Run("loads skeleton from repository", func(t *testing.T) {
		repo := createTestMultiRepo(t, "repo1")

		ref, err := repo.CreateSkeleton("repo1:myskeleton")
		require.NoError(t, err)

		skeleton, err := repo.LoadSkeleton("repo1:myskeleton")
		require.NoError(t, err)

		require.Equal(t, ref, skeleton.Ref)
		assert.Len(t, skeleton.Files, 1)
	})

	t.Run("returns error if skeleton does not exist", func(t *testing.T) {
		repo := createTestMultiRepo(t, "repo1")

		_, err := repo.LoadSkeleton("repo1:nonexistent")
		require.Error(t, err)
	})
}

func createTestMultiRepo(t *testing.T, repos ...string) kickoff.Repository {
	repoMap := make(map[string]string, len(repos))

	for _, name := range repos {
		dir := filepath.Join(t.TempDir(), name)

		_, err := Create(dir)
		require.NoError(t, err)

		repoMap[name] = dir
	}

	repo, err := OpenMap(context.Background(), repoMap, nil)
	require.NoError(t, err)

	return repo
}
