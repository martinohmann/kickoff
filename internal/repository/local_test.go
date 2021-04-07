package repository

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalRepository_GetSkeleton(t *testing.T) {
	info := kickoff.RepoRef{Path: "../testdata/repos/repo1"}

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
		ref := kickoff.RepoRef{Path: "../testdata/repos/repo1"}

		repo, err := NewLocalRepository(ref)
		require.NoError(t, err)

		infos, err := repo.ListSkeletons(context.Background())
		require.NoError(t, err)

		require.Len(t, infos, 2)
		assert.Equal(t, "advanced", infos[0].Name)
		assert.Equal(t, "minimal", infos[1].Name)
	})

	t.Run("listing nonexistent repository causes error", func(t *testing.T) {
		repo, err := NewLocalRepository(kickoff.RepoRef{})
		require.NoError(t, err)

		_, err = repo.ListSkeletons(context.Background())
		require.Error(t, err)
	})
}

func TestFindSkeletons(t *testing.T) {
	t.Run("finds all skeletons", func(t *testing.T) {
		ref := &kickoff.RepoRef{Path: "../testdata/repos/advanced"}

		skeletons, err := findSkeletons(ref, filepath.Join(ref.Path, "skeletons"))
		require.NoError(t, err)

		pwd, err := os.Getwd()
		require.NoError(t, err)

		path := func(name string) string {
			return filepath.Join(pwd, ref.Path, "skeletons", name)
		}

		expected := []*kickoff.SkeletonRef{
			{Name: "bar", Path: path("bar"), Repo: ref},
			{Name: "child", Path: path("child"), Repo: ref},
			{Name: "childofchild", Path: path("childofchild"), Repo: ref},
			{Name: "cyclea", Path: path("cyclea"), Repo: ref},
			{Name: "cycleb", Path: path("cycleb"), Repo: ref},
			{Name: "cyclec", Path: path("cyclec"), Repo: ref},
			{Name: "foo/bar", Path: path("foo/bar"), Repo: ref},
			{Name: "nested/dir", Path: path("nested/dir"), Repo: ref},
			{Name: "parent", Path: path("parent"), Repo: ref},
		}

		require.Equal(t, expected, skeletons)
	})

	t.Run("FindSkeletons returns error if RepoInfo points to nonexistent dir", func(t *testing.T) {
		ref := &kickoff.RepoRef{Path: "../nonexistent"}
		_, err := findSkeletons(ref, filepath.Join(ref.Path, "skeletons"))
		require.Error(t, err)
	})
}
