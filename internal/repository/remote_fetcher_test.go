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
	"github.com/martinohmann/kickoff/internal/git"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDefaultFetcher_FetchRemote(t *testing.T) {
	t.Run("local refs are a no-op", func(t *testing.T) {
		fetcher := NewRemoteFetcher(nil)
		ref := kickoff.RepoRef{Path: t.TempDir()}

		require.NoError(t, fetcher.FetchRemote(context.Background(), ref))
	})

	t.Run("it opens a local repository and checks out the correct revision", func(t *testing.T) {
		defer testutil.MockRepositoryCacheDir(t.TempDir())()
		fetcher, fakeClient := newTestRemoteFetcher()
		ref, localPath := newTestRepoRef()

		createLocalTestRepoDir(t, localPath, time.Now())

		fakeRepo := &git.FakeRepository{}

		fakeClient.On("Open", localPath).Return(fakeRepo, nil)

		hash := plumbing.NewHash("de4db3ef")

		fakeRepo.On("ResolveRevision", plumbing.Revision("master")).Return(&hash, nil)
		fakeRepo.On("Checkout", hash).Return(nil)

		require.NoError(t, fetcher.FetchRemote(context.Background(), ref))
	})

	t.Run("it opens a local repository, fetches refs and checks out the correct revision", func(t *testing.T) {
		defer testutil.MockRepositoryCacheDir(t.TempDir())()
		fetcher, fakeClient := newTestRemoteFetcher()
		ref, localPath := newTestRepoRef()

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

		require.NoError(t, fetcher.FetchRemote(ctx, ref))
	})

	t.Run("it clones a remote repository if not present in local dir", func(t *testing.T) {
		defer testutil.MockRepositoryCacheDir(t.TempDir())()
		fetcher, fakeClient := newTestRemoteFetcher()
		ref, localPath := newTestRepoRef()

		fakeRepo := &git.FakeRepository{}

		ctx := context.Background()

		fakeClient.On("Open", localPath).Return(nil, git.ErrRepositoryNotExists)
		fakeClient.On("Clone", ctx, "https://git.kickoff.tld/owner/repo", localPath).
			Run(func(args mock.Arguments) {
				// we are simulating cloning by just creating a new skeleton repository.
				createLocalTestRepoDir(t, localPath, time.Now())
			}).
			Return(fakeRepo, nil)

		hash := plumbing.NewHash("de4db3ef")

		fakeRepo.On("Fetch", ctx, []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"}).Return(git.NoErrAlreadyUpToDate)
		fakeRepo.On("ResolveRevision", plumbing.Revision("master")).Return(&hash, nil)
		fakeRepo.On("Checkout", hash).Return(nil)

		require.NoError(t, fetcher.FetchRemote(ctx, ref))
	})

	t.Run("returns error if open returns error different from git.ErrRepositoryNotExists", func(t *testing.T) {
		defer testutil.MockRepositoryCacheDir(t.TempDir())()
		fetcher, fakeClient := newTestRemoteFetcher()
		ref, localPath := newTestRepoRef()

		ctx := context.Background()

		openErr := errors.New("open failed")

		fakeClient.On("Open", localPath).Return(nil, openErr)

		err := fetcher.FetchRemote(ctx, ref)
		require.Equal(t, openErr, err)
	})

	t.Run("returns error if clone fails", func(t *testing.T) {
		defer testutil.MockRepositoryCacheDir(t.TempDir())()
		fetcher, fakeClient := newTestRemoteFetcher()
		ref, localPath := newTestRepoRef()

		ctx := context.Background()

		cloneErr := errors.New("clone failed")

		fakeClient.On("Open", localPath).Return(nil, git.ErrRepositoryNotExists)
		fakeClient.On("Clone", ctx, "https://git.kickoff.tld/owner/repo", localPath).Return(nil, cloneErr)

		err := fetcher.FetchRemote(ctx, ref)
		require.Equal(t, cloneErr, err)
	})

	t.Run("returns error if fetch fails", func(t *testing.T) {
		defer testutil.MockRepositoryCacheDir(t.TempDir())()
		fetcher, fakeClient := newTestRemoteFetcher()
		ref, localPath := newTestRepoRef()

		// simulate a cached repository that hasn't been modified since 10
		// minutes.
		createLocalTestRepoDir(t, localPath, time.Now().Add(-10*time.Minute))

		fakeRepo := &git.FakeRepository{}

		fakeClient.On("Open", localPath).Return(fakeRepo, nil)

		ctx := context.Background()

		fetchErr := errors.New("fetch failed")

		fakeRepo.On("Fetch", ctx, []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"}).Return(fetchErr)

		err := fetcher.FetchRemote(ctx, ref)
		require.Equal(t, fetchErr, err)
	})

	t.Run("handles temporary network errors", func(t *testing.T) {
		defer testutil.MockRepositoryCacheDir(t.TempDir())()
		fetcher, fakeClient := newTestRemoteFetcher()
		ref, localPath := newTestRepoRef()

		// simulate a cached repository that hasn't been modified since 10
		// minutes.
		createLocalTestRepoDir(t, localPath, time.Now().Add(-10*time.Minute))

		fakeRepo := &git.FakeRepository{}

		fakeClient.On("Open", localPath).Return(fakeRepo, nil)

		ctx := context.Background()

		fakeRepo.On("Fetch", ctx, []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"}).
			Return(&net.DNSError{IsTemporary: true})

		require.NoError(t, fetcher.FetchRemote(ctx, ref))
	})

	t.Run("cleans up after git reference error", func(t *testing.T) {
		defer testutil.MockRepositoryCacheDir(t.TempDir())()
		fetcher, fakeClient := newTestRemoteFetcher()
		ref, localPath := newTestRepoRef()

		createLocalTestRepoDir(t, localPath, time.Now())

		fakeRepo := &git.FakeRepository{}

		fakeClient.On("Open", localPath).Return(fakeRepo, nil)

		fakeRepo.On("ResolveRevision", plumbing.Revision("master")).
			Return(nil, plumbing.ErrReferenceNotFound)
		fakeRepo.On("ResolveRevision", plumbing.Revision(plumbing.NewTagReferenceName("master"))).
			Return(nil, plumbing.ErrReferenceNotFound)
		fakeRepo.On("ResolveRevision", plumbing.Revision(plumbing.NewRemoteReferenceName("origin", "master"))).
			Return(nil, plumbing.ErrReferenceNotFound)

		err := fetcher.FetchRemote(context.Background(), ref)
		require.EqualError(t, err, `revision "master" not found in repository "https://git.kickoff.tld/owner/repo"`)

		require.NoDirExists(t, localPath)
	})

	t.Run("attempts to resolve multiple revisions", func(t *testing.T) {
		defer testutil.MockRepositoryCacheDir(t.TempDir())()
		fetcher, fakeClient := newTestRemoteFetcher()
		ref, localPath := newTestRepoRef()

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

		require.NoError(t, fetcher.FetchRemote(context.Background(), ref))
	})

	t.Run("does not checkout revision if empty", func(t *testing.T) {
		defer testutil.MockRepositoryCacheDir(t.TempDir())()
		fetcher, fakeClient := newTestRemoteFetcher()
		ref, localPath := newTestRepoRef("")

		createLocalTestRepoDir(t, localPath, time.Now())

		fakeRepo := &git.FakeRepository{}

		fakeClient.On("Open", localPath).Return(fakeRepo, nil)

		require.NoError(t, fetcher.FetchRemote(context.Background(), ref))
	})

	t.Run("it propagates context to (git.Client).Clone", func(t *testing.T) {
		defer testutil.MockRepositoryCacheDir(t.TempDir())()
		fetcher, fakeClient := newTestRemoteFetcher()
		ref, localPath := newTestRepoRef()

		createLocalTestRepoDir(t, localPath, time.Now())

		fakeClient.On("Open", localPath).Return(nil, git.ErrRepositoryNotExists)

		mockCall := fakeClient.On("Clone", mock.Anything, "https://git.kickoff.tld/owner/repo", localPath)
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

		err := fetcher.FetchRemote(ctx, ref)
		require.Error(t, err)
		require.Same(t, context.Canceled, err)
	})

	t.Run("it propagates context to (git.Repository).Fetch", func(t *testing.T) {
		defer testutil.MockRepositoryCacheDir(t.TempDir())()
		fetcher, fakeClient := newTestRemoteFetcher()
		ref, localPath := newTestRepoRef()

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

		err := fetcher.FetchRemote(ctx, ref)
		require.Error(t, err)
		require.Same(t, context.Canceled, err)
	})
}

func createLocalTestRepoDir(t *testing.T, dir string, modTime time.Time) {
	repo, err := Create(dir)
	require.NoError(t, err)
	_, err = repo.CreateSkeleton("default")
	require.NoError(t, err)
	require.NoError(t, os.Chtimes(dir, modTime, modTime))
}

func newTestRemoteFetcher() (RemoteFetcher, *git.FakeClient) {
	fakeClient := &git.FakeClient{}
	fetcher := NewRemoteFetcher(fakeClient)
	return fetcher, fakeClient
}

func newTestRepoRef(revision ...string) (kickoff.RepoRef, string) {
	rev := "master"
	if len(revision) > 0 {
		rev = revision[0]
	}

	ref := kickoff.RepoRef{
		URL:      "https://git.kickoff.tld/owner/repo",
		Revision: rev,
	}

	return ref, ref.LocalPath()
}
