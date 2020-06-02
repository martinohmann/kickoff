package skeleton

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/martinohmann/kickoff/internal/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInfo(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	t.Run("string", func(t *testing.T) {
		info := &Info{Name: "default"}
		assert.Equal("default", info.String())
		info = &Info{Name: "default", Repo: &RepoInfo{Name: "the-repo"}}
		assert.Equal("the-repo:default", info.String())
	})

	t.Run("load skeleton config from info", func(t *testing.T) {
		info := &Info{
			Name: "default",
			Path: "../testdata/repos/repo1/skeletons/minimal",
		}

		config, err := info.LoadConfig()
		require.NoError(err)

		expectedConfig := Config{
			Values: template.Values{"foo": "bar"},
		}

		assert.Equal(expectedConfig, config)
	})
}

func TestRepoInfo(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	t.Run("is remote?", func(t *testing.T) {
		info := &RepoInfo{Name: "default"}
		assert.False(info.IsRemote())
		info = &RepoInfo{Name: "default", Path: "some/path"}
		assert.False(info.IsRemote())
		info = &RepoInfo{Name: "default", Path: "some/path", URL: "https://some/remote/url"}
		assert.True(info.IsRemote())
	})

	t.Run("finds all skeletons", func(t *testing.T) {
		info := &RepoInfo{Path: "../testdata/repos/advanced"}

		skeletons, err := info.FindSkeletons()
		require.NoError(err)

		pwd, err := os.Getwd()
		require.NoError(err)

		path := func(name string) string {
			return filepath.Join(pwd, info.Path, "skeletons", name)
		}

		expected := []*Info{
			{Name: "bar", Path: path("bar"), Repo: info},
			{Name: "child", Path: path("child"), Repo: info},
			{Name: "childofchild", Path: path("childofchild"), Repo: info},
			{Name: "cyclea", Path: path("cyclea"), Repo: info},
			{Name: "cycleb", Path: path("cycleb"), Repo: info},
			{Name: "cyclec", Path: path("cyclec"), Repo: info},
			{Name: "foo/bar", Path: path("foo/bar"), Repo: info},
			{Name: "nested/dir", Path: path("nested/dir"), Repo: info},
			{Name: "parent", Path: path("parent"), Repo: info},
		}

		require.Equal(expected, skeletons)
	})

	t.Run("FindSkeletons returns error if RepoInfo points to nonexistent dir", func(t *testing.T) {
		info := &RepoInfo{Path: "../nonexistent"}
		_, err := info.FindSkeletons()
		require.Error(err)
	})
}
