package repository

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/martinohmann/kickoff/internal/file"
	"github.com/martinohmann/kickoff/internal/git"
	"github.com/martinohmann/kickoff/internal/skeleton"
	log "github.com/sirupsen/logrus"
)

// RemoteRepository is a skeleton repository that resides in a remote git
// repository.
type RemoteRepository struct {
	*LocalRepository
	revision string
	url      string

	err    error
	once   sync.Once
	client git.Client
}

// NewRemoteRepository creates a *RemoteRepository from info. Returns
// ErrNotARemoteRepository if info does not describe a remote repository
// location. Internally creates a *LocalRepository for the cached copy of the
// remote and returns any error that might occur while creating it.
func NewRemoteRepository(info skeleton.RepoInfo) (*RemoteRepository, error) {
	if !info.IsRemote() {
		return nil, ErrNotARemoteRepository
	}

	local, err := NewLocalRepository(info)
	if err != nil {
		return nil, err
	}

	r := &RemoteRepository{
		LocalRepository: local,
		url:             info.URL,
		revision:        info.Revision,
		client:          git.NewClient(),
	}

	return r, nil
}

// GetSkeleton implements Repository.
//
// Lazily synchronizes the cached local copy of the remote repository before
// looking up the skeleton.
func (r *RemoteRepository) GetSkeleton(ctx context.Context, name string) (*skeleton.Info, error) {
	err := r.syncRemoteOnce(ctx)
	if err != nil {
		return nil, err
	}

	return r.LocalRepository.GetSkeleton(ctx, name)
}

// ListSkeletons implements Repository.
//
// Lazily synchronizes the cached local copy of the remote repository before
// listing skeletons.
func (r *RemoteRepository) ListSkeletons(ctx context.Context) ([]*skeleton.Info, error) {
	err := r.syncRemoteOnce(ctx)
	if err != nil {
		return nil, err
	}

	return r.LocalRepository.ListSkeletons(ctx)
}

func (r *RemoteRepository) syncRemoteOnce(ctx context.Context) error {
	r.once.Do(func() {
		r.err = r.syncRemote(ctx)
	})

	return r.err
}

func (r *RemoteRepository) syncRemote(ctx context.Context) error {
	localPath := r.info.Path

	err := r.updateLocalCache(ctx, r.url, r.revision)
	if err == nil {
		return nil
	}

	if !file.Exists(localPath) {
		return err
	}

	var netErr net.Error

	if errors.As(err, &netErr) && netErr.Temporary() {
		// If we have the repository in the local cache and this is a network
		// error, we will do best effort and try to serve the local cache
		// instead of failing. At least log a warning.
		log.WithError(netErr).
			WithField("url", r.url).
			Warn("failed to update local repository cache")

		return nil
	}

	if errors.Is(err, plumbing.ErrReferenceNotFound) {
		// A git reference error indicates that we cloned a repository but
		// the desired revision was not found. The local cache is in a
		// potentially invalid state now and needs to be cleaned.
		log.WithField("path", localPath).Debug("cleaning up repository cache")

		osErr := os.RemoveAll(localPath)
		if osErr != nil {
			err = fmt.Errorf("failed to cleanup after error: %v: %w", err, osErr)
		}
	}

	return err
}

func (r *RemoteRepository) updateLocalCache(ctx context.Context, url, revision string) error {
	repo, err := r.fetchOrCloneRemote(ctx, url)
	if err != nil {
		return err
	}

	return checkoutRevision(repo, revision)
}

func (r *RemoteRepository) fetchOrCloneRemote(ctx context.Context, url string) (git.Repository, error) {
	localPath := r.info.Path

	repo, err := r.client.Open(localPath)
	if err == git.ErrRepositoryNotExists {
		log.WithFields(log.Fields{
			"url":  url,
			"path": localPath,
		}).Debug("cloning remote repository")

		return r.client.Clone(ctx, url, localPath)
	} else if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{"path": localPath}).Debug("opened repository")

	err = r.fetchRefs(ctx, repo)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *RemoteRepository) fetchRefs(ctx context.Context, repo git.Repository) error {
	localPath := r.info.Path

	// As git operations that fetch refs from remotes can be slow, we are
	// trying to avoid doing too many of them. We are going to only attempt
	// to fetch refs if the modification timestamp of the checked out local
	// repo is older than one minute. This should be a reasonable time
	// frame that is long enough to see a noticeable speed up of issuing
	// multiple commands that read from repositories (e.g. `kickoff
	// skeleton list`, `kickoff skeleton show foobar`) shortly after
	// another, but it is short enough to avoid having a stale version of a
	// remote repository checked out locally for too long.
	fileInfo, err := os.Stat(localPath)
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

	log.WithFields(log.Fields{"refs": refs}).Debug("fetching refs")

	err = repo.Fetch(ctx, refs...)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	now = time.Now()

	// Important: update the modification date after fetching the refs so
	// we can actually make use of the improvement above.
	return os.Chtimes(localPath, now, now)
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
