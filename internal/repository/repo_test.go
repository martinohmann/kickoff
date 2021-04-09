package repository

import (
	"context"
	"testing"

	"github.com/martinohmann/kickoff/internal/kickoff"
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
		assert.IsType(t, &localRepository{}, repo)
	})

	t.Run("creates remote repositories", func(t *testing.T) {
		repo, err := New("https://github.com/martinohmann/kickoff-skeletons")
		require.NoError(t, err)
		assert.IsType(t, &remoteRepository{}, repo)
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

func TestNewFromRef_Invalid(t *testing.T) {
	invalidRef := kickoff.RepoRef{Path: "/foo/bar", URL: "https://foo.bar.baz"}

	_, err := NewFromRef(invalidRef)
	require.EqualError(t, err, "invalid repository ref: URL and Path must not be set at the same time")
}
