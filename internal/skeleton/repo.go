package skeleton

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/apex/log"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/martinohmann/kickoff/internal/file"
)

// Repository is the interface for a skeleton repository.
type Repository interface {
	// SkeletonInfo obtains the info for the skeleton with name or an error if the
	// skeleton does not exist within the repository.
	SkeletonInfo(name string) (*Info, error)

	// SkeletonInfos returns infos for all skeletons available in the repository.
	// Returns any error that may occur while traversing the directory.
	SkeletonInfos() ([]*Info, error)
}

// repoCacheKey uniquely identifies a locally configured repository by name and
// url. Used as a key for the repository cache below.
type repoCacheKey struct {
	name, url string
}

var (
	// repoCache is used to avoid unnecessary git and fs operations for
	// repositories that have already been initialized.
	repoCache map[repoCacheKey]Repository

	// repoCacheMu protects repoCache
	repoCacheMu sync.Mutex
)

// OpenRepository opens a repository and returns it. If url points to a remote
// repository it will be looked up in the local cache and reused if possible.
// If the repository is not in the cache it will be cloned. Open will
// automatically checks out the revision provided in the url. Returns any
// errors that occur while parsing the url opening the repository directory or
// during git actions.
func OpenRepository(url string) (Repository, error) {
	return openNamedRepository("", url)
}

func openNamedRepository(name, url string) (repo Repository, err error) {
	cacheKey := repoCacheKey{name, url}

	repoCacheMu.Lock()
	defer repoCacheMu.Unlock()

	if repo, ok := repoCache[cacheKey]; ok {
		return repo, nil
	}

	info, err := ParseRepositoryURL(url)
	if err != nil {
		return nil, err
	}

	info.Name = name

	if info.Local {
		repo, err = openLocalRepository(info)
	} else {
		repo, err = openRemoteRepository(info)
	}

	if err != nil {
		return nil, err
	}

	err = info.Validate()
	if err != nil {
		return nil, err
	}

	if repoCache == nil {
		repoCache = make(map[repoCacheKey]Repository)
	}

	repoCache[cacheKey] = repo

	return repo, nil
}

func openLocalRepository(info *RepositoryInfo) (Repository, error) {
	path := info.LocalPath()

	ok, err := file.IsDirectory(path)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, fmt.Errorf("%s is not a directory", path)
	}

	return &repository{info}, nil
}

func openRemoteRepository(info *RepositoryInfo) (Repository, error) {
	err := updateLocalGitRepository(info)
	if err == nil {
		return &repository{info}, nil
	}

	if file.Exists(info.LocalPath()) {
		if netErr, ok := err.(net.Error); ok {
			// If we have the repository in the local cache and this is a network
			// error, we will do best effort and try to serve the local cache
			// instead of failing. At least log a warning.
			log.WithField("url", info.String()).
				Warnf("falling back to potentially stale repository from local cache due to network error: %v", netErr)

			return openLocalRepository(info)
		}

		if err == plumbing.ErrReferenceNotFound {
			// A git reference error indicates that we cloned a repository but
			// the desired revision was not found. The local cache is in a
			// potentially invalid state now and needs to be cleaned.
			log.WithField("path", info.LocalPath()).Debug("cleaning up repository cache")

			os.RemoveAll(info.LocalPath())
		}
	}

	return nil, err
}

func updateLocalGitRepository(info *RepositoryInfo) error {
	log.WithFields(log.Fields{
		"url":   info.String(),
		"local": info.LocalPath(),
	}).Debug("using remote skeleton repository")

	repo, err := openOrCloneGitRepository(info)
	if err != nil {
		return err
	}

	hash, err := resolveRevision(repo, info.Revision)
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	log.WithField("sha1", hash.String()).Debug("checking out commit")

	return worktree.Checkout(&git.CheckoutOptions{
		Hash:  *hash,
		Force: true,
	})
}

func openOrCloneGitRepository(info *RepositoryInfo) (*git.Repository, error) {
	repo, err := git.PlainOpen(info.LocalPath())
	switch {
	case err == git.ErrRepositoryNotExists:
		return git.PlainClone(info.LocalPath(), false, &git.CloneOptions{
			URL: info.String(),
		})
	case err != nil:
		return nil, err
	default:
		err := repo.Fetch(&git.FetchOptions{
			RefSpecs: []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"},
		})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return nil, err
		}

		return repo, nil
	}
}

func resolveRevision(repo *git.Repository, revision string) (*plumbing.Hash, error) {
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
			"sha1":     hash.String(),
			"revision": rev,
		}).Debug("resolved revision")

		return hash, nil
	}

	return nil, plumbing.ErrReferenceNotFound
}

type repository struct {
	info *RepositoryInfo
}

// SkeletonInfo implements Repository.
func (r *repository) SkeletonInfo(name string) (*Info, error) {
	path := filepath.Join(r.info.SkeletonsDir(), name)

	ok, err := isSkeletonDir(path)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, fmt.Errorf("skeleton %q not found in %s", name, r.info)
	}

	info := &Info{
		Name: name,
		Path: path,
		Repo: r.info,
	}

	return info, nil
}

// SkeletonInfos implements Repository.
func (r *repository) SkeletonInfos() ([]*Info, error) {
	return findSkeletons(r.info, r.info.SkeletonsDir())
}

// findSkeletons recursively finds all skeletons in dir. Returns any error that
// may occur while traversing dir.
func findSkeletons(repo *RepositoryInfo, dir string) ([]*Info, error) {
	skeletons := make([]*Info, 0)

	fileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, info := range fileInfos {
		if !info.IsDir() {
			continue
		}

		path := filepath.Join(dir, info.Name())

		ok, err := isSkeletonDir(path)
		if os.IsPermission(err) {
			log.Warnf("permission error, skipping dir: %v", err)
			continue
		}

		if err != nil {
			return nil, err
		}

		if ok {
			abspath, err := filepath.Abs(path)
			if err != nil {
				return nil, err
			}

			skeletons = append(skeletons, &Info{
				Name: info.Name(),
				Path: abspath,
				Repo: repo,
			})
			continue
		}

		skels, err := findSkeletons(repo, path)
		if err != nil {
			return nil, err
		}

		for _, s := range skels {
			skeletons = append(skeletons, &Info{
				Name: filepath.Join(info.Name(), s.Name),
				Path: s.Path,
				Repo: repo,
			})
		}
	}

	return skeletons, nil
}
