package repository

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalRepository_GetSkeleton(t *testing.T) {
	info := kickoff.RepoRef{Path: "../testdata/repos/repo1"}

	repo := newLocal(info)

	abspath, err := filepath.Abs("../testdata/repos/repo1")
	require.NoError(t, err)

	t.Run("can retrieve info about a single skeleton", func(t *testing.T) {
		info, err := repo.GetSkeleton(context.Background(), "minimal")
		require.NoError(t, err)

		assert.Equal(t, abspath, info.Repo.LocalPath())
		assert.Equal(t, filepath.Join(abspath, kickoff.SkeletonsDir, "minimal"), info.Path)
		assert.Equal(t, "minimal", info.Name)
	})

	t.Run("nonexistent repository causes error", func(t *testing.T) {
		repo := newLocal(kickoff.RepoRef{Path: "/non/existent/repo"})

		_, err := repo.GetSkeleton(context.Background(), "default")
		require.EqualError(t, err, `"/non/existent/repo" is not a valid skeleton repository`)
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

		repo := newLocal(ref)

		infos, err := repo.ListSkeletons(context.Background())
		require.NoError(t, err)

		require.Len(t, infos, 2)
		assert.Equal(t, "advanced", infos[0].Name)
		assert.Equal(t, "minimal", infos[1].Name)
	})

	t.Run("listing nonexistent repository causes error", func(t *testing.T) {
		repo := newLocal(kickoff.RepoRef{})

		_, err := repo.ListSkeletons(context.Background())
		require.Error(t, err)
	})
}

func TestFindSkeletons(t *testing.T) {
	t.Run("finds all skeletons", func(t *testing.T) {
		ref := &kickoff.RepoRef{Path: "../testdata/repos/advanced"}

		skeletons, err := findSkeletons(ref, ref.SkeletonsPath())
		require.NoError(t, err)

		expected := []*kickoff.SkeletonRef{
			{Name: "bar", Path: ref.SkeletonPath("bar"), Repo: ref},
			{Name: "child", Path: ref.SkeletonPath("child"), Repo: ref},
			{Name: "childofchild", Path: ref.SkeletonPath("childofchild"), Repo: ref},
			{Name: "cyclea", Path: ref.SkeletonPath("cyclea"), Repo: ref},
			{Name: "cycleb", Path: ref.SkeletonPath("cycleb"), Repo: ref},
			{Name: "cyclec", Path: ref.SkeletonPath("cyclec"), Repo: ref},
			{Name: "foo/bar", Path: ref.SkeletonPath("foo/bar"), Repo: ref},
			{Name: "nested/dir", Path: ref.SkeletonPath("nested/dir"), Repo: ref},
			{Name: "parent", Path: ref.SkeletonPath("parent"), Repo: ref},
		}

		require.Equal(t, expected, skeletons)
	})

	t.Run("FindSkeletons returns error if RepoInfo points to nonexistent dir", func(t *testing.T) {
		ref := &kickoff.RepoRef{Path: "../nonexistent"}
		_, err := findSkeletons(ref, ref.SkeletonsPath())
		require.Error(t, err)
	})
}
