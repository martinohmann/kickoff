package kickoff

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRepoRef(t *testing.T) {
	// override local user cache dir to be able to make test assertions on
	// paths.
	oldCacheDir := LocalRepositoryCacheDir
	LocalRepositoryCacheDir = "/home/someuser/.cache/kickoff/repositories"
	defer func() { LocalRepositoryCacheDir = oldCacheDir }()

	t.Run("parses absolute paths", func(t *testing.T) {
		ref, err := ParseRepoRef("/some/absolute/path")
		require.NoError(t, err)
		require.False(t, ref.IsRemote())
		assert.Equal(t, "/some/absolute/path", ref.Path)
	})

	t.Run("parses relative paths", func(t *testing.T) {
		ref, err := ParseRepoRef("../some/relative/path")
		require.NoError(t, err)
		require.False(t, ref.IsRemote())
		assert.Equal(t, "../some/relative/path", ref.Path)
	})

	t.Run("parses absolute paths", func(t *testing.T) {
		ref, err := ParseRepoRef("/some/absolute/path")
		require.NoError(t, err)
		require.False(t, ref.IsRemote())
		assert.Equal(t, "/some/absolute/path", ref.Path)
	})

	t.Run("parses homedir paths", func(t *testing.T) {
		os.Setenv("HOME", "/home/user")
		ref, err := ParseRepoRef("~/repo")
		require.NoError(t, err)
		require.False(t, ref.IsRemote())
		assert.Equal(t, "/home/user/repo", ref.Path)
	})

	t.Run("parses remote urls", func(t *testing.T) {
		ref, err := ParseRepoRef("https://example.com/some/repo")
		require.NoError(t, err)
		require.True(t, ref.IsRemote())
		assert.Equal(t, "/home/someuser/.cache/kickoff/repositories/example.com/some/repo@master", ref.Path)
		assert.Equal(t, "https://example.com/some/repo", ref.URL)
		assert.Equal(t, "master", ref.Revision)
	})

	t.Run("parses revision from remote urls", func(t *testing.T) {
		ref, err := ParseRepoRef("https://example.com/some/repo?revision=de4db3ef")
		require.NoError(t, err)
		require.True(t, ref.IsRemote())
		assert.Equal(t, "/home/someuser/.cache/kickoff/repositories/example.com/some/repo@de4db3ef", ref.Path)
		assert.Equal(t, "https://example.com/some/repo", ref.URL)
		assert.Equal(t, "de4db3ef", ref.Revision)
	})

	t.Run("returns errors on invalid query", func(t *testing.T) {
		_, err := ParseRepoRef("https://example.com/some/repo?%")
		require.Error(t, err)
	})
}
