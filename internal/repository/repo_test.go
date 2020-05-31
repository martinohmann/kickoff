package repository

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	// disable caching
	oldCache := repoCache
	defer func() { repoCache = oldCache }()
	DisableCache()

	t.Run("creates local repositories", func(t *testing.T) {
		repo, err := New("../testdata/repos/repo1")
		require.NoError(t, err)
		assert.IsType(t, &LocalRepository{}, repo)
	})

	t.Run("creates remote repositories", func(t *testing.T) {
		repo, err := New("https://github.com/martinohmann/kickoff-skeletons")
		require.NoError(t, err)
		assert.IsType(t, &RemoteRepository{}, repo)
	})

	t.Run("enabled repository cache", func(t *testing.T) {
		EnableCache()

		repo1, err := New("../testdata/repos/repo1")
		require.NoError(t, err)

		repo2, err := New("../testdata/repos/repo1")
		require.NoError(t, err)

		assert.Same(t, repo1, repo2)
	})

	t.Run("disabled repository cache", func(t *testing.T) {
		DisableCache()

		repo1, err := New("../testdata/repos/repo1")
		require.NoError(t, err)

		repo2, err := New("../testdata/repos/repo1")
		require.NoError(t, err)

		if repo1 == repo2 {
			t.Fatal("pointer are equal when they should not")
		}
	})

	t.Run("fails to create repositories from invalid urls", func(t *testing.T) {
		_, err := New("\nhttpxd::/asdf\\invalid")
		require.Error(t, err)
	})
}

func TestNewNamed(t *testing.T) {
	// disable caching
	oldCache := repoCache
	defer func() { repoCache = oldCache }()
	DisableCache()

	t.Run("propagates name into skeleton info", func(t *testing.T) {
		repo, err := NewNamed("the-name", "../testdata/repos/repo1")
		require.NoError(t, err)
		assert.NotNil(t, repo)

		info, err := repo.GetSkeleton(context.Background(), "minimal")
		require.NoError(t, err)
		assert.Equal(t, "the-name", info.Repo.Name)
	})
}

func TestParseURL(t *testing.T) {
	// override local user cache dir to be able to make test assertions on
	// paths.
	oldCacheDir := LocalCache
	LocalCache = "/home/someuser/.cache/kickoff/repositories"
	defer func() { LocalCache = oldCacheDir }()

	t.Run("parses absolute paths", func(t *testing.T) {
		info, err := ParseURL("/some/absolute/path")
		require.NoError(t, err)
		require.False(t, info.IsRemote())
		assert.Equal(t, "/some/absolute/path", info.Path)
	})

	t.Run("parses relative paths", func(t *testing.T) {
		info, err := ParseURL("../some/relative/path")
		require.NoError(t, err)
		require.False(t, info.IsRemote())
		assert.Equal(t, "../some/relative/path", info.Path)
	})

	t.Run("parses absolute paths", func(t *testing.T) {
		info, err := ParseURL("/some/absolute/path")
		require.NoError(t, err)
		require.False(t, info.IsRemote())
		assert.Equal(t, "/some/absolute/path", info.Path)
	})

	t.Run("parses homedir paths", func(t *testing.T) {
		os.Setenv("HOME", "/home/user")
		info, err := ParseURL("~/repo")
		require.NoError(t, err)
		require.False(t, info.IsRemote())
		assert.Equal(t, "/home/user/repo", info.Path)
	})

	t.Run("parses remote urls", func(t *testing.T) {
		info, err := ParseURL("https://example.com/some/repo")
		require.NoError(t, err)
		require.True(t, info.IsRemote())
		assert.Equal(t, "/home/someuser/.cache/kickoff/repositories/example.com/some/repo@master", info.Path)
		assert.Equal(t, "https://example.com/some/repo", info.URL)
		assert.Equal(t, "master", info.Revision)
	})

	t.Run("parses revision from remote urls", func(t *testing.T) {
		info, err := ParseURL("https://example.com/some/repo?revision=de4db3ef")
		require.NoError(t, err)
		require.True(t, info.IsRemote())
		assert.Equal(t, "/home/someuser/.cache/kickoff/repositories/example.com/some/repo@de4db3ef", info.Path)
		assert.Equal(t, "https://example.com/some/repo", info.URL)
		assert.Equal(t, "de4db3ef", info.Revision)
	})

	t.Run("returns errors on invalid query", func(t *testing.T) {
		_, err := ParseURL("https://example.com/some/repo?%")
		require.Error(t, err)
	})
}
