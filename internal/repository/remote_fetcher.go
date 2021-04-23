package repository

import (
	"context"
	"errors"
	"net"
	"os"
	"time"

	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/martinohmann/kickoff/internal/git"
	"github.com/martinohmann/kickoff/internal/kickoff"
	log "github.com/sirupsen/logrus"
)

// RemoteFetcher can fetch remote repositories.
type RemoteFetcher interface {
	// FetchRemote fetches the remote repository referenced by ref and places
	// it in the local path dictated by the ref and checks out the desired
	// revision (if configured).
	// Must return non-recoverable errors while fetching the repository or
	// checking out the desired revision.
	FetchRemote(ctx context.Context, ref kickoff.RepoRef) error
}

var defaultFetcher = NewRemoteFetcher(git.NewClient())

// remoteFetcher is a skeleton repository that resides in a remote git
// repository.
type remoteFetcher struct {
	client git.Client
}

// NewRemoteFetcher creates a RemoteFetcher which uses a git client to clone
// and fetch remote repositories.
func NewRemoteFetcher(client git.Client) RemoteFetcher {
	return &remoteFetcher{client: client}
}

func (r *remoteFetcher) FetchRemote(ctx context.Context, ref kickoff.RepoRef) error {
	if ref.IsLocal() {
		return nil
	}

	err := r.updateLocalCache(ctx, ref)
	if err == nil {
		return nil
	}

	localPath := ref.LocalPath()

	if _, statErr := os.Stat(localPath); statErr != nil {
		return err
	}

	var netErr net.Error

	if errors.As(err, &netErr) && netErr.Temporary() {
		// If we have the repository in the local cache and this is a network
		// error, we will do best effort and try to serve the local cache
		// instead of failing. At least log a warning.
		log.WithError(netErr).
			WithField("url", ref.URL).
			Warn("failed to update local repository cache")

		return nil
	}

	if errors.Is(err, plumbing.ErrReferenceNotFound) {
		err = RevisionNotFoundError{RepoRef: ref}
		// A git reference error indicates that we cloned a repository but
		// the desired revision was not found. The local cache is in a
		// potentially invalid state now and needs to be cleaned.
		log.WithField("path", localPath).Debug("cleaning up repository cache")

		if err := os.RemoveAll(localPath); err != nil {
			log.WithError(err).
				WithField("path", localPath).
				Error("failed to cleanup cache dir")
		}
	}

	return err
}

func (r *remoteFetcher) updateLocalCache(ctx context.Context, ref kickoff.RepoRef) error {
	repo, err := r.fetchOrCloneRemote(ctx, ref.URL, ref.LocalPath())
	if err != nil {
		return err
	}

	if ref.Revision == "" {
		return nil
	}

	return checkoutRevision(repo, ref.Revision)
}

func (r *remoteFetcher) fetchOrCloneRemote(ctx context.Context, url, path string) (git.Repository, error) {
	repo, err := r.client.Open(path)
	if err == git.ErrRepositoryNotExists {
		log.WithFields(log.Fields{
			"url":  url,
			"path": path,
		}).Debug("cloning remote repository")

		return r.client.Clone(ctx, url, path)
	} else if err != nil {
		return nil, err
	}

	log.WithField("path", path).Debug("opened repository")

	err = r.fetchRefs(ctx, repo, path)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *remoteFetcher) fetchRefs(ctx context.Context, repo git.Repository, path string) error {
	// As git operations that fetch refs from remotes can be slow, we are
	// trying to avoid doing too many of them. We are going to only attempt
	// to fetch refs if the modification timestamp of the checked out local
	// repo is older than one minute. This should be a reasonable time
	// frame that is long enough to see a noticeable speed up of issuing
	// multiple commands that read from repositories (e.g. `kickoff
	// skeleton list`, `kickoff skeleton show foobar`) shortly after
	// another, but it is short enough to avoid having a stale version of a
	// remote repository checked out locally for too long.
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}

	now := time.Now()
	modTime := fileInfo.ModTime()

	// @TODO(mohmann): maybe make this configurable?
	if modTime.Add(1 * time.Minute).After(now) {
		// Refs were already fetched less than a minute ago, bail out early
		// as there will probably be nothing new most of the time.
		return nil
	}

	refs := []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"}

	log.WithField("refs", refs).Debug("fetching refs")

	err = repo.Fetch(ctx, refs...)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	now = time.Now()

	// Important: update the modification date after fetching the refs so
	// we can actually make use of the improvement above.
	return os.Chtimes(path, now, now)
}

func resolveRevision(repo git.Repository, revision string) (*plumbing.Hash, error) {
	revisions := []plumbing.Revision{
		plumbing.Revision(revision),
		plumbing.Revision(plumbing.NewTagReferenceName(revision)),
		plumbing.Revision(plumbing.NewRemoteReferenceName("origin", revision)),
	}

	for _, rev := range revisions {
		hash, err := repo.ResolveRevision(rev)
		if err == plumbing.ErrReferenceNotFound {
			continue
		}

		if err != nil {
			return nil, err
		}

		log.WithFields(log.Fields{
			"hash":     hash.String(),
			"revision": rev,
		}).Debug("resolved revision")

		return hash, nil
	}

	return nil, plumbing.ErrReferenceNotFound
}

func checkoutRevision(repo git.Repository, revision string) error {
	hash, err := resolveRevision(repo, revision)
	if err != nil {
		return err
	}

	log.WithField("hash", hash.String()).Debug("checking out commit")

	return repo.Checkout(*hash)
}
