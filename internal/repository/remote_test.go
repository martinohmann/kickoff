package repository

import (
	"context"
	"errors"
	"io/ioutil"
	"net"
	"os"
	"testing"
	"time"

	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/martinohmann/kickoff/internal/file"
	"github.com/martinohmann/kickoff/internal/git"
	"github.com/martinohmann/kickoff/internal/skeleton"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewRemoteRepository(t *testing.T) {
	t.Run("creates remote repository", func(t *testing.T) {
		repo, err := NewRemoteRepository(skeleton.RepoInfo{
			Path:     "some/local/path",
			URL:      "https://github.com/martinohmann/kickoff-skeletons",
			Revision: "de4db3ef",
		})
		require.NoError(t, err)
		require.NotNil(t, repo)
	})

	t.Run("returns error if info does not describe a remote repo", func(t *testing.T) {
		_, err := NewRemoteRepository(skeleton.RepoInfo{Path: "some/local/path"})
		assert.Equal(t, ErrNotARemoteRepository, err)
	})
}

func TestRemoteRepository_syncRemote(t *testing.T) {
	require := require.New(t)

	t.Run("it opens a local repository and checks out the correct revision", func(t *testing.T) {
		repo, fakeClient, tmpdir, cleanup := createRemoteTestRepo(t)
		defer cleanup()

		fakeRepo := &git.FakeRepository{}

		fakeClient.On("Open", tmpdir).Run(func(args mock.Arguments) {
			createLocalTestRepoDir(t, tmpdir, time.Now())
		}).Return(fakeRepo, nil)

		hash := plumbing.NewHash("de4db3ef")

		fakeRepo.On("ResolveRevision", plumbing.Revision("master")).Return(&hash, nil)
		fakeRepo.On("Checkout", hash).Return(nil)

		require.NoError(repo.syncRemote(context.Background()))
	})

	t.Run("it opens a local repository, fetches refs and checks out the correct revision", func(t *testing.T) {
		repo, fakeClient, tmpdir, cleanup := createRemoteTestRepo(t)
		defer cleanup()

		fakeRepo := &git.FakeRepository{}

		fakeClient.On("Open", tmpdir).Run(func(args mock.Arguments) {
			// simulate a cached repository that hasn't been modified since 10
			// minutes.
			createLocalTestRepoDir(t, tmpdir, time.Now().Add(-10*time.Minute))
		}).Return(fakeRepo, nil)

		hash := plumbing.NewHash("de4db3ef")

		ctx := context.Background()

		fakeRepo.On("Fetch", ctx, []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"}).Return(nil)
		fakeRepo.On("ResolveRevision", plumbing.Revision("master")).Return(&hash, nil)
		fakeRepo.On("Checkout", hash).Return(nil)

		require.NoError(repo.syncRemote(ctx))
	})

	t.Run("it clones a remote repository if not present in local dir", func(t *testing.T) {
		repo, fakeClient, tmpdir, cleanup := createRemoteTestRepo(t)
		defer cleanup()

		fakeRepo := &git.FakeRepository{}

		ctx := context.Background()

		fakeClient.On("Open", tmpdir).Return(nil, git.ErrRepositoryNotExists)
		fakeClient.On("Clone", ctx, "https://github.com/martinohmann/kickoff-skeletons", tmpdir).
			Run(func(args mock.Arguments) { createLocalTestRepoDir(t, tmpdir, time.Now()) }).
			Return(fakeRepo, nil)

		hash := plumbing.NewHash("de4db3ef")

		fakeRepo.On("Fetch", ctx, []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"}).Return(git.NoErrAlreadyUpToDate)
		fakeRepo.On("ResolveRevision", plumbing.Revision("master")).Return(&hash, nil)
		fakeRepo.On("Checkout", hash).Return(nil)

		require.NoError(repo.syncRemote(ctx))
	})

	t.Run("returns error if open returns error different from git.ErrRepositoryNotExists", func(t *testing.T) {
		repo, fakeClient, tmpdir, cleanup := createRemoteTestRepo(t)
		defer cleanup()

		ctx := context.Background()

		openErr := errors.New("open failed")

		fakeClient.On("Open", tmpdir).Return(nil, openErr)

		err := repo.syncRemote(ctx)
		require.Equal(openErr, err)
	})

	t.Run("returns error if clone fails", func(t *testing.T) {
		repo, fakeClient, tmpdir, cleanup := createRemoteTestRepo(t)
		defer cleanup()

		ctx := context.Background()

		cloneErr := errors.New("clone failed")

		fakeClient.On("Open", tmpdir).Return(nil, git.ErrRepositoryNotExists)
		fakeClient.On("Clone", ctx, "https://github.com/martinohmann/kickoff-skeletons", tmpdir).Return(nil, cloneErr)

		err := repo.syncRemote(ctx)
		require.Equal(cloneErr, err)
	})

	t.Run("returns error if fetch fails", func(t *testing.T) {
		repo, fakeClient, tmpdir, cleanup := createRemoteTestRepo(t)
		defer cleanup()

		fakeRepo := &git.FakeRepository{}

		fakeClient.On("Open", tmpdir).Run(func(args mock.Arguments) {
			// simulate a cached repository that hasn't been modified since 10
			// minutes.
			createLocalTestRepoDir(t, tmpdir, time.Now().Add(-10*time.Minute))
		}).Return(fakeRepo, nil)

		ctx := context.Background()

		fetchErr := errors.New("fetch failed")

		fakeRepo.On("Fetch", ctx, []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"}).Return(fetchErr)

		err := repo.syncRemote(ctx)
		require.Equal(fetchErr, err)
	})

	t.Run("handles temporary network errors", func(t *testing.T) {
		repo, fakeClient, tmpdir, cleanup := createRemoteTestRepo(t)
		defer cleanup()

		fakeRepo := &git.FakeRepository{}

		fakeClient.On("Open", tmpdir).Run(func(args mock.Arguments) {
			// simulate a cached repository that hasn't been modified since 10
			// minutes.
			createLocalTestRepoDir(t, tmpdir, time.Now().Add(-10*time.Minute))
		}).Return(fakeRepo, nil)

		ctx := context.Background()

		fakeRepo.On("Fetch", ctx, []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"}).
			Return(&net.DNSError{IsTemporary: true})

		require.NoError(repo.syncRemote(ctx))
	})

	t.Run("cleans up after git reference error", func(t *testing.T) {
		repo, fakeClient, tmpdir, cleanup := createRemoteTestRepo(t)
		defer cleanup()

		fakeRepo := &git.FakeRepository{}

		fakeClient.On("Open", tmpdir).Run(func(args mock.Arguments) {
			createLocalTestRepoDir(t, tmpdir, time.Now())
		}).Return(fakeRepo, nil)

		fakeRepo.On("ResolveRevision", plumbing.Revision("master")).
			Return(nil, plumbing.ErrReferenceNotFound)
		fakeRepo.On("ResolveRevision", plumbing.Revision(plumbing.NewTagReferenceName("master"))).
			Return(nil, plumbing.ErrReferenceNotFound)
		fakeRepo.On("ResolveRevision", plumbing.Revision(plumbing.NewRemoteReferenceName("origin", "master"))).
			Return(nil, plumbing.ErrReferenceNotFound)

		err := repo.syncRemote(context.Background())
		require.Equal(plumbing.ErrReferenceNotFound, err)

		require.False(file.Exists(tmpdir))
	})

	t.Run("attempts to resolve multiple revisions", func(t *testing.T) {
		repo, fakeClient, tmpdir, cleanup := createRemoteTestRepo(t)
		defer cleanup()

		fakeRepo := &git.FakeRepository{}

		fakeClient.On("Open", tmpdir).Run(func(args mock.Arguments) {
			createLocalTestRepoDir(t, tmpdir, time.Now())
		}).Return(fakeRepo, nil)

		hash := plumbing.NewHash("de4db3ef")

		fakeRepo.On("ResolveRevision", plumbing.Revision("master")).
			Return(nil, plumbing.ErrReferenceNotFound)
		fakeRepo.On("ResolveRevision", plumbing.Revision(plumbing.NewTagReferenceName("master"))).
			Return(nil, plumbing.ErrReferenceNotFound)
		fakeRepo.On("ResolveRevision", plumbing.Revision(plumbing.NewRemoteReferenceName("origin", "master"))).
			Return(&hash, nil)
		fakeRepo.On("Checkout", hash).Return(nil)

		require.NoError(repo.syncRemote(context.Background()))
	})
}

func TestRemoteRepository_GetSkeleton(t *testing.T) {
	require := require.New(t)

	t.Run("retrieves skeleton info from local repo copy", func(t *testing.T) {
		repo, fakeClient, tmpdir, cleanup := createRemoteTestRepo(t)
		defer cleanup()

		fakeRepo := &git.FakeRepository{}

		fakeClient.On("Open", tmpdir).Run(func(args mock.Arguments) {
			createLocalTestRepoDir(t, tmpdir, time.Now())
		}).Return(fakeRepo, nil)

		hash := plumbing.NewHash("de4db3ef")

		fakeRepo.On("ResolveRevision", plumbing.Revision("master")).Return(&hash, nil)
		fakeRepo.On("Checkout", hash).Return(nil)

		info, err := repo.GetSkeleton(context.Background(), "default")
		require.NoError(err)
		require.Equal("default", info.Name)

		_, err = repo.GetSkeleton(context.Background(), "nonexistent")
		require.Error(err)
		require.IsType(SkeletonNotFoundError{}, err)
	})

	t.Run("fails if repo sync fails", func(t *testing.T) {
		repo, fakeClient, tmpdir, cleanup := createRemoteTestRepo(t)
		defer cleanup()

		openErr := errors.New("open failed")

		fakeClient.On("Open", tmpdir).Return(nil, openErr)

		_, err := repo.GetSkeleton(context.Background(), "nonexistent")
		require.Equal(openErr, err)
	})

	t.Run("syncs only once", func(t *testing.T) {
		repo, fakeClient, tmpdir, cleanup := createRemoteTestRepo(t)
		defer cleanup()

		openErr := errors.New("open failed")

		// We expect only one call to Open.
		fakeClient.On("Open", tmpdir).Once().Return(nil, openErr)

		// Consecutive calls to GetSkeleton must return the same error.
		_, err := repo.GetSkeleton(context.Background(), "default")
		require.Equal(openErr, err)
		_, err = repo.GetSkeleton(context.Background(), "nonexistent")
		require.Equal(openErr, err)
	})
}

func TestRemoteRepository_ListSkeletons(t *testing.T) {
	require := require.New(t)

	t.Run("lists skeleton infos from local repo copy", func(t *testing.T) {
		repo, fakeClient, tmpdir, cleanup := createRemoteTestRepo(t)
		defer cleanup()

		fakeRepo := &git.FakeRepository{}

		fakeClient.On("Open", tmpdir).Run(func(args mock.Arguments) {
			createLocalTestRepoDir(t, tmpdir, time.Now())
		}).Return(fakeRepo, nil)

		hash := plumbing.NewHash("de4db3ef")

		fakeRepo.On("ResolveRevision", plumbing.Revision("master")).Return(&hash, nil)
		fakeRepo.On("Checkout", hash).Return(nil)

		infos, err := repo.ListSkeletons(context.Background())
		require.NoError(err)
		require.Len(infos, 1)
		require.Equal("default", infos[0].Name)
	})

	t.Run("fails if repo sync fails", func(t *testing.T) {
		repo, fakeClient, tmpdir, cleanup := createRemoteTestRepo(t)
		defer cleanup()

		openErr := errors.New("open failed")

		fakeClient.On("Open", tmpdir).Return(nil, openErr)

		_, err := repo.ListSkeletons(context.Background())
		require.Equal(openErr, err)
	})
}

func createLocalTestRepoDir(t *testing.T, dir string, modTime time.Time) {
	require.NoError(t, Create(dir, "default"))
	require.NoError(t, os.Chtimes(dir, modTime, modTime))
}

func createRemoteTestRepo(t *testing.T) (*RemoteRepository, *git.FakeClient, string, func()) {
	tmpdir, err := ioutil.TempDir("", "kickoff-repos-*")
	require.NoError(t, err)

	repo, err := NewRemoteRepository(skeleton.RepoInfo{
		Path:     tmpdir,
		URL:      "https://github.com/martinohmann/kickoff-skeletons",
		Revision: "master",
	})
	require.NoError(t, err)

	fakeClient := &git.FakeClient{}
	repo.client = fakeClient

	return repo, fakeClient, tmpdir, func() { os.RemoveAll(tmpdir) }
}
