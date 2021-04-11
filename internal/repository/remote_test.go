package repository

import (
	"context"
	"errors"
	"net"
	"os"
	"testing"
	"time"

	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/martinohmann/kickoff/internal/file"
	"github.com/martinohmann/kickoff/internal/git"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewRemoteRepository(t *testing.T) {
	defer testutil.MockRepositoryCacheDir(t.TempDir())()

	t.Run("creates remote repository", func(t *testing.T) {
		repo, err := newRemote(kickoff.RepoRef{
			URL:      "https://github.com/martinohmann/kickoff-skeletons",
			Revision: "de4db3ef",
		})
		require.NoError(t, err)
		require.NotNil(t, repo)
	})

	t.Run("returns error if ref does not describe a remote repo", func(t *testing.T) {
		_, err := newRemote(kickoff.RepoRef{Path: "some/local/path"})
		assert.Equal(t, ErrNotARemoteRepository, err)
	})
}

func TestRemoteRepository_syncRemote(t *testing.T) {
	t.Run("it opens a local repository and checks out the correct revision", func(t *testing.T) {
		repo, fakeClient, localPath, restore := createRemoteTestRepo(t)
		defer restore()

		createLocalTestRepoDir(t, localPath, time.Now())

		fakeRepo := &git.FakeRepository{}

		fakeClient.On("Open", localPath).Return(fakeRepo, nil)

		hash := plumbing.NewHash("de4db3ef")

		fakeRepo.On("ResolveRevision", plumbing.Revision("master")).Return(&hash, nil)
		fakeRepo.On("Checkout", hash).Return(nil)

		require.NoError(t, repo.syncRemote(context.Background()))
	})

	t.Run("it opens a local repository, fetches refs and checks out the correct revision", func(t *testing.T) {
		repo, fakeClient, localPath, restore := createRemoteTestRepo(t)
		defer restore()

		// simulate a cached repository that hasn't been modified since 10
		// minutes.
		createLocalTestRepoDir(t, localPath, time.Now().Add(-10*time.Minute))

		fakeRepo := &git.FakeRepository{}

		fakeClient.On("Open", localPath).Return(fakeRepo, nil)

		hash := plumbing.NewHash("de4db3ef")

		ctx := context.Background()

		fakeRepo.On("Fetch", ctx, []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"}).Return(nil)
		fakeRepo.On("ResolveRevision", plumbing.Revision("master")).Return(&hash, nil)
		fakeRepo.On("Checkout", hash).Return(nil)

		require.NoError(t, repo.syncRemote(ctx))
	})

	t.Run("it clones a remote repository if not present in local dir", func(t *testing.T) {
		repo, fakeClient, localPath, restore := createRemoteTestRepo(t)
		defer restore()

		fakeRepo := &git.FakeRepository{}

		ctx := context.Background()

		fakeClient.On("Open", localPath).Return(nil, git.ErrRepositoryNotExists)
		fakeClient.On("Clone", ctx, "https://github.com/martinohmann/kickoff-skeletons", localPath).
			Run(func(args mock.Arguments) {
				// we are simulating cloning by just creating a new skeleton repository.
				createLocalTestRepoDir(t, localPath, time.Now())
			}).
			Return(fakeRepo, nil)

		hash := plumbing.NewHash("de4db3ef")

		fakeRepo.On("Fetch", ctx, []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"}).Return(git.NoErrAlreadyUpToDate)
		fakeRepo.On("ResolveRevision", plumbing.Revision("master")).Return(&hash, nil)
		fakeRepo.On("Checkout", hash).Return(nil)

		require.NoError(t, repo.syncRemote(ctx))
	})

	t.Run("returns error if open returns error different from git.ErrRepositoryNotExists", func(t *testing.T) {
		repo, fakeClient, localPath, restore := createRemoteTestRepo(t)
		defer restore()

		ctx := context.Background()

		openErr := errors.New("open failed")

		fakeClient.On("Open", localPath).Return(nil, openErr)

		err := repo.syncRemote(ctx)
		require.Equal(t, openErr, err)
	})

	t.Run("returns error if clone fails", func(t *testing.T) {
		repo, fakeClient, localPath, restore := createRemoteTestRepo(t)
		defer restore()

		ctx := context.Background()

		cloneErr := errors.New("clone failed")

		fakeClient.On("Open", localPath).Return(nil, git.ErrRepositoryNotExists)
		fakeClient.On("Clone", ctx, "https://github.com/martinohmann/kickoff-skeletons", localPath).Return(nil, cloneErr)

		err := repo.syncRemote(ctx)
		require.Equal(t, cloneErr, err)
	})

	t.Run("returns error if fetch fails", func(t *testing.T) {
		repo, fakeClient, localPath, restore := createRemoteTestRepo(t)
		defer restore()

		// simulate a cached repository that hasn't been modified since 10
		// minutes.
		createLocalTestRepoDir(t, localPath, time.Now().Add(-10*time.Minute))

		fakeRepo := &git.FakeRepository{}

		fakeClient.On("Open", localPath).Return(fakeRepo, nil)

		ctx := context.Background()

		fetchErr := errors.New("fetch failed")

		fakeRepo.On("Fetch", ctx, []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"}).Return(fetchErr)

		err := repo.syncRemote(ctx)
		require.Equal(t, fetchErr, err)
	})

	t.Run("handles temporary network errors", func(t *testing.T) {
		repo, fakeClient, localPath, restore := createRemoteTestRepo(t)
		defer restore()

		// simulate a cached repository that hasn't been modified since 10
		// minutes.
		createLocalTestRepoDir(t, localPath, time.Now().Add(-10*time.Minute))

		fakeRepo := &git.FakeRepository{}

		fakeClient.On("Open", localPath).Return(fakeRepo, nil)

		ctx := context.Background()

		fakeRepo.On("Fetch", ctx, []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"}).
			Return(&net.DNSError{IsTemporary: true})

		require.NoError(t, repo.syncRemote(ctx))
	})

	t.Run("cleans up after git reference error", func(t *testing.T) {
		repo, fakeClient, localPath, restore := createRemoteTestRepo(t)
		defer restore()

		createLocalTestRepoDir(t, localPath, time.Now())

		fakeRepo := &git.FakeRepository{}

		fakeClient.On("Open", localPath).Return(fakeRepo, nil)

		fakeRepo.On("ResolveRevision", plumbing.Revision("master")).
			Return(nil, plumbing.ErrReferenceNotFound)
		fakeRepo.On("ResolveRevision", plumbing.Revision(plumbing.NewTagReferenceName("master"))).
			Return(nil, plumbing.ErrReferenceNotFound)
		fakeRepo.On("ResolveRevision", plumbing.Revision(plumbing.NewRemoteReferenceName("origin", "master"))).
			Return(nil, plumbing.ErrReferenceNotFound)

		err := repo.syncRemote(context.Background())
		require.EqualError(t, err, `revision "master" not found in repository "https://github.com/martinohmann/kickoff-skeletons"`)

		require.False(t, file.Exists(localPath))
	})

	t.Run("attempts to resolve multiple revisions", func(t *testing.T) {
		repo, fakeClient, localPath, restore := createRemoteTestRepo(t)
		defer restore()

		createLocalTestRepoDir(t, localPath, time.Now())

		fakeRepo := &git.FakeRepository{}

		fakeClient.On("Open", localPath).Return(fakeRepo, nil)

		hash := plumbing.NewHash("de4db3ef")

		fakeRepo.On("ResolveRevision", plumbing.Revision("master")).
			Return(nil, plumbing.ErrReferenceNotFound)
		fakeRepo.On("ResolveRevision", plumbing.Revision(plumbing.NewTagReferenceName("master"))).
			Return(nil, plumbing.ErrReferenceNotFound)
		fakeRepo.On("ResolveRevision", plumbing.Revision(plumbing.NewRemoteReferenceName("origin", "master"))).
			Return(&hash, nil)
		fakeRepo.On("Checkout", hash).Return(nil)

		require.NoError(t, repo.syncRemote(context.Background()))
	})

	t.Run("does not checkout revision if empty", func(t *testing.T) {
		repo, fakeClient, localPath, restore := createRemoteTestRepo(t, "")
		defer restore()

		createLocalTestRepoDir(t, localPath, time.Now())

		fakeRepo := &git.FakeRepository{}

		fakeClient.On("Open", localPath).Return(fakeRepo, nil)

		require.NoError(t, repo.syncRemote(context.Background()))
	})

	t.Run("it propagates context to (git.Client).Clone", func(t *testing.T) {
		repo, fakeClient, localPath, restore := createRemoteTestRepo(t)
		defer restore()

		createLocalTestRepoDir(t, localPath, time.Now())

		fakeClient.On("Open", localPath).Return(nil, git.ErrRepositoryNotExists)

		mockCall := fakeClient.On("Clone", mock.Anything, "https://github.com/martinohmann/kickoff-skeletons", localPath)
		mockCall.RunFn = func(args mock.Arguments) {
			ctx := args.Get(0).(context.Context)
			select {
			case <-ctx.Done():
				mockCall.ReturnArguments = mock.Arguments{nil, ctx.Err()}
			default:
				mockCall.ReturnArguments = mock.Arguments{nil, errors.New("context not propagated")}
			}
		}

		ctx, cancel := context.WithCancel(context.Background())
		// cancel the context immediately
		cancel()

		err := repo.syncRemote(ctx)
		require.Error(t, err)
		require.Same(t, context.Canceled, err)
	})

	t.Run("it propagates context to (git.Repository).Fetch", func(t *testing.T) {
		repo, fakeClient, localPath, restore := createRemoteTestRepo(t)
		defer restore()

		// simulate a cached repository that hasn't been modified since 10
		// minutes.
		createLocalTestRepoDir(t, localPath, time.Now().Add(-10*time.Minute))

		fakeRepo := &git.FakeRepository{}

		fakeClient.On("Open", localPath).Return(fakeRepo, nil)

		mockCall := fakeRepo.On("Fetch", mock.Anything, []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"})
		mockCall.RunFn = func(args mock.Arguments) {
			ctx := args.Get(0).(context.Context)
			select {
			case <-ctx.Done():
				mockCall.ReturnArguments = mock.Arguments{ctx.Err()}
			default:
				mockCall.ReturnArguments = mock.Arguments{errors.New("context not propagated")}
			}
		}

		ctx, cancel := context.WithCancel(context.Background())
		// cancel the context immediately
		cancel()

		err := repo.syncRemote(ctx)
		require.Error(t, err)
		require.Same(t, context.Canceled, err)
	})
}

func TestRemoteRepository_GetSkeleton(t *testing.T) {
	t.Run("retrieves skeleton ref from local repo copy", func(t *testing.T) {
		repo, fakeClient, localPath, restore := createRemoteTestRepo(t)
		defer restore()

		createLocalTestRepoDir(t, localPath, time.Now())

		fakeRepo := &git.FakeRepository{}

		fakeClient.On("Open", localPath).Return(fakeRepo, nil)

		hash := plumbing.NewHash("de4db3ef")

		fakeRepo.On("ResolveRevision", plumbing.Revision("master")).Return(&hash, nil)
		fakeRepo.On("Checkout", hash).Return(nil)

		ref, err := repo.GetSkeleton(context.Background(), "default")
		require.NoError(t, err)
		require.Equal(t, "default", ref.Name)

		_, err = repo.GetSkeleton(context.Background(), "nonexistent")
		require.Error(t, err)
		require.IsType(t, SkeletonNotFoundError{}, err)
	})

	t.Run("fails if repo sync fails", func(t *testing.T) {
		repo, fakeClient, localPath, restore := createRemoteTestRepo(t)
		defer restore()

		openErr := errors.New("open failed")

		fakeClient.On("Open", localPath).Return(nil, openErr)

		_, err := repo.GetSkeleton(context.Background(), "nonexistent")
		require.Equal(t, openErr, err)
	})

	t.Run("syncs only once", func(t *testing.T) {
		repo, fakeClient, localPath, restore := createRemoteTestRepo(t)
		defer restore()

		openErr := errors.New("open failed")

		// We expect only one call to Open.
		fakeClient.On("Open", localPath).Once().Return(nil, openErr)

		// Consecutive calls to GetSkeleton must return the same error.
		_, err := repo.GetSkeleton(context.Background(), "default")
		require.Equal(t, openErr, err)
		_, err = repo.GetSkeleton(context.Background(), "nonexistent")
		require.Equal(t, openErr, err)
	})
}

func TestRemoteRepository_ListSkeletons(t *testing.T) {
	t.Run("lists skeleton refs from local repo copy", func(t *testing.T) {
		repo, fakeClient, localPath, restore := createRemoteTestRepo(t)
		defer restore()

		createLocalTestRepoDir(t, localPath, time.Now())

		fakeRepo := &git.FakeRepository{}

		fakeClient.On("Open", localPath).Return(fakeRepo, nil)

		hash := plumbing.NewHash("de4db3ef")

		fakeRepo.On("ResolveRevision", plumbing.Revision("master")).Return(&hash, nil)
		fakeRepo.On("Checkout", hash).Return(nil)

		refs, err := repo.ListSkeletons(context.Background())
		require.NoError(t, err)
		require.Len(t, refs, 1)
		require.Equal(t, "default", refs[0].Name)
	})

	t.Run("fails if repo sync fails", func(t *testing.T) {
		repo, fakeClient, localPath, restore := createRemoteTestRepo(t)
		defer restore()

		openErr := errors.New("open failed")

		fakeClient.On("Open", localPath).Return(nil, openErr)

		_, err := repo.ListSkeletons(context.Background())
		require.Equal(t, openErr, err)
	})
}

func createLocalTestRepoDir(t *testing.T, dir string, modTime time.Time) {
	require.NoError(t, CreateWithSkeleton(dir, "default"))
	require.NoError(t, os.Chtimes(dir, modTime, modTime))
}

func createRemoteTestRepo(t *testing.T, revision ...string) (*remoteRepository, *git.FakeClient, string, func()) {
	rev := "master"
	if len(revision) > 0 {
		rev = revision[0]
	}

	restoreCacheDir := testutil.MockRepositoryCacheDir(t.TempDir())
	ref := kickoff.RepoRef{
		URL:      "https://github.com/martinohmann/kickoff-skeletons",
		Revision: rev,
	}

	repo, err := newRemote(ref)
	require.NoError(t, err)

	fakeClient := &git.FakeClient{}
	repo.client = fakeClient

	return repo, fakeClient, ref.LocalPath(), restoreCacheDir
}
