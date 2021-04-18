package repository

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/template"
	"github.com/martinohmann/kickoff/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeFetcher struct {
	fn  func(ref kickoff.RepoRef) error
	err error
}

func (f *fakeFetcher) FetchRemote(ctx context.Context, ref kickoff.RepoRef) error {
	if f.fn != nil {
		return f.fn(ref)
	}
	return f.err
}

func TestOpen(t *testing.T) {
	t.Run("nonexistent repository causes error", func(t *testing.T) {
		_, err := Open(context.Background(), "/non/existent/repo", nil)
		require.EqualError(t, err, `"/non/existent/repo" is not a valid skeleton repository`)
	})

	t.Run("opens local repositories", func(t *testing.T) {
		_, err := Open(context.Background(), "../testdata/repos/repo1", nil)
		require.NoError(t, err)
	})

	t.Run("opens remote repositories", func(t *testing.T) {
		defer testutil.MockRepositoryCacheDir(t.TempDir())()

		_, err := Open(context.Background(), "https://foo.bar/baz/qux", &Options{
			Fetcher: &fakeFetcher{fn: func(ref kickoff.RepoRef) error {
				return os.MkdirAll(ref.SkeletonsPath(), 0755)
			}},
		})
		require.NoError(t, err)
	})

	t.Run("returns error if fetching remote fails", func(t *testing.T) {
		_, err := Open(context.Background(), "https://github.com/martinohmann/kickoff-skeletons", &Options{
			Fetcher: &fakeFetcher{err: errors.New("failed to fetch remote")},
		})
		require.Error(t, err)
	})

	t.Run("fails to create repositories from invalid urls", func(t *testing.T) {
		_, err := Open(context.Background(), "\nhttpxd::/asdf\\invalid", nil)
		require.Error(t, err)
	})
}

func TestOpenNamed(t *testing.T) {
	t.Run("propagates name into skeleton ref", func(t *testing.T) {
		repo, err := openNamed(context.Background(), "the-name", "../testdata/repos/repo1", nil)
		require.NoError(t, err)

		ref, err := repo.GetSkeleton("minimal")
		require.NoError(t, err)
		assert.Equal(t, "the-name", ref.Repo.Name)
	})
}

func TestOpenRef(t *testing.T) {
	t.Run("validates passed in ref", func(t *testing.T) {
		invalidRef := kickoff.RepoRef{Path: "/foo/bar", URL: "https://foo.bar.baz"}

		_, err := OpenRef(context.Background(), invalidRef, nil)
		require.EqualError(t, err, "invalid repository ref: URL and Path must not be set at the same time")
	})
}

func TestRepository_GetSkeleton(t *testing.T) {
	t.Parallel()

	ref := kickoff.RepoRef{Name: "the-repo", Path: "../testdata/repos/repo1"}

	repo, err := newRepository(ref)
	require.NoError(t, err)

	abspath, err := filepath.Abs("../testdata/repos/repo1")
	require.NoError(t, err)

	t.Run("can retrieve a single skeleton", func(t *testing.T) {
		ref, err := repo.GetSkeleton("minimal")
		require.NoError(t, err)

		assert.Equal(t, abspath, ref.Repo.LocalPath())
		assert.Equal(t, filepath.Join(abspath, kickoff.SkeletonsDir, "minimal"), ref.Path)
		assert.Equal(t, "minimal", ref.Name)
	})

	t.Run("returns SkeletonNotFoundError if skeleton does not exist", func(t *testing.T) {
		_, err := repo.GetSkeleton("nonexistent")
		require.EqualError(t, err, `skeleton "nonexistent" not found in repository "the-repo"`)
	})
}

func TestRepository_ListSkeletons(t *testing.T) {
	t.Run("can list all skeletons", func(t *testing.T) {
		ref := kickoff.RepoRef{Path: "../testdata/repos/repo1"}

		repo, err := newRepository(ref)
		require.NoError(t, err)

		refs, err := repo.ListSkeletons()
		require.NoError(t, err)

		require.Len(t, refs, 2)
		assert.Equal(t, "advanced", refs[0].Name)
		assert.Equal(t, "minimal", refs[1].Name)
	})
}

func TestRepository_CreateSkeleton(t *testing.T) {
	t.Parallel()

	t.Run("create skeleton in existing local repository", func(t *testing.T) {
		repo, err := Create(t.TempDir() + "/repo")
		require.NoError(t, err)

		ref, err := repo.CreateSkeleton("myskeleton")
		require.NoError(t, err)

		skeletonPath := ref.Repo.SkeletonPath("myskeleton")

		require.FileExists(t, filepath.Join(skeletonPath, "README.md.skel"))
		require.FileExists(t, filepath.Join(skeletonPath, kickoff.SkeletonConfigFileName))
	})

	t.Run("cannot create skeleton with empty name", func(t *testing.T) {
		repo, err := Create(t.TempDir() + "/repo")
		require.NoError(t, err)

		_, err = repo.CreateSkeleton("")
		require.Error(t, err)
	})

	t.Run("cannot overwrite existing skeleton", func(t *testing.T) {
		repo, err := Create(t.TempDir() + "/repo")
		require.NoError(t, err)

		_, err = repo.CreateSkeleton("myskeleton")
		require.NoError(t, err)

		_, err = repo.CreateSkeleton("myskeleton")
		require.EqualError(t, err, `skeleton "myskeleton" already exists`)
	})

	t.Run("cannot create skeleton in remote repository", func(t *testing.T) {
		defer testutil.MockRepositoryCacheDir(t.TempDir())()

		repo, err := Open(context.Background(), "https://remote.tld/owner/repo", &Options{
			Fetcher: &fakeFetcher{fn: func(ref kickoff.RepoRef) error {
				return os.MkdirAll(ref.SkeletonsPath(), 0755)
			}},
		})
		require.NoError(t, err)

		_, err = repo.CreateSkeleton("myskeleton")
		require.EqualError(t, err, "creating skeletons in remote repositories is not supported")
	})
}

func TestListSkeletons(t *testing.T) {
	t.Parallel()

	t.Run("lists all skeletons", func(t *testing.T) {
		ref := &kickoff.RepoRef{Path: "../testdata/repos/advanced"}
		refs, err := listSkeletons(ref, ref.SkeletonsPath())
		require.NoError(t, err)

		expected := []*kickoff.SkeletonRef{
			{Name: "bar", Path: ref.SkeletonPath("bar"), Repo: ref},
			{Name: "foo/bar", Path: ref.SkeletonPath("foo/bar"), Repo: ref},
			{Name: "nested/dir", Path: ref.SkeletonPath("nested/dir"), Repo: ref},
		}

		require.Equal(t, expected, refs)
	})

	t.Run("returns error if ref points to nonexistent dir", func(t *testing.T) {
		ref := &kickoff.RepoRef{Path: "../nonexistent"}
		_, err := listSkeletons(ref, ref.SkeletonsPath())
		require.Error(t, err)
	})
}

func TestLoadSkeletons(t *testing.T) {
	ref := kickoff.RepoRef{
		Name: "the-repo",
		Path: "../testdata/repos/repo1",
	}

	repo, err := OpenRef(context.Background(), ref, nil)
	require.NoError(t, err)

	skeletons, err := LoadSkeletons(repo, []string{"advanced"})
	require.NoError(t, err)

	require.Len(t, skeletons, 1)

	skeleton := skeletons[0]

	assert.Equal(t, "advanced", skeleton.Ref.Name)
	assert.Equal(t, &ref, skeleton.Ref.Repo)
	assert.Len(t, skeleton.Files, 4)
	assert.Equal(t,
		template.Values{
			"somekey":         "somevalue",
			"filename":        "foobar",
			"optionalContent": "",
			"travis": map[string]interface{}{
				"enabled": false,
			},
		},
		skeleton.Values,
	)
}
