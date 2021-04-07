package repository

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/kirsle/configdir"
	"github.com/martinohmann/kickoff/internal/homedir"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/skeleton"
)

// LocalCache holds the path to the local repository cache. This is platform
// specific.
var LocalCache = configdir.LocalCache("kickoff", "repositories")

// Repository is the interface for a skeleton repository.
type Repository interface {
	// GetSkeleton retrieves information about a skeleton from the repository.
	// The passed in context is propagated to all operations that cross API
	// boundaries (e.g. git operations) and can be used to enforce timeouts or
	// cancel them. Returns an error of type SkeletonNotFoundError if the named
	// skeleton was not found in the repository.
	GetSkeleton(ctx context.Context, name string) (*skeleton.Info, error)

	// ListSkeletons retrieves information about all skeletons in the
	// repository. The passed in context is propagated to all operations that
	// cross API boundaries (e.g. git operations) and can be used to enforce
	// timeouts or cancel them. If the repository is empty, ListSkeletons will
	// return an empty slice.
	ListSkeletons(ctx context.Context) ([]*skeleton.Info, error)
}

// New creates a new Repository for url. Returns an error if url is not a valid
// local path or remote url.
func New(url string) (Repository, error) {
	return NewNamed("", url)
}

// NewNamed creates a new named Repository. The name is propagated into the
// repository info that is attached to every skeleton that is retrieved from
// it. Apart from that is behaves exactly like New.
func NewNamed(name, url string) (repo Repository, err error) {
	key := cacheKey{name, url}

	if repo, ok := repoCache.get(key); ok {
		return repo, nil
	}

	info, err := ParseURL(url)
	if err != nil {
		return nil, err
	}

	info.Name = name

	if info.IsRemote() {
		repo, err = NewRemoteRepository(*info)
	} else {
		repo, err = NewLocalRepository(*info)
	}

	if err != nil {
		return nil, err
	}

	repoCache.set(key, repo)

	return repo, nil
}

// ParseURL parses a raw repository url and returns a repository info
// describing a local or remote skeleton repository. The rawurl parameter must
// be either a local path or a remote url to a git repository. Remote url may
// optionally include a `revision` query parameter. If absent, `master` will be
// assumed. Returns an error if url does not match any of the criteria
// mentioned above.
func ParseURL(rawurl string) (*kickoff.RepoRef, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, fmt.Errorf("invalid repo URL %q: %w", rawurl, err)
	}

	if u.Host == "" {
		path, err := homedir.Expand(u.Path)
		if err != nil {
			return nil, err
		}

		return &kickoff.RepoRef{Path: path}, nil
	}

	query, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return nil, fmt.Errorf("invalid URL query %q: %w", u.RawQuery, err)
	}

	var revision string
	if rev, ok := query["revision"]; ok && len(rev) > 0 {
		revision = rev[0]
	}

	if revision == "" {
		revision = "master"
	}

	// Query is only used to pass an optional revision and needs to be empty in
	// the final repository URL.
	u.RawQuery = ""

	return &kickoff.RepoRef{
		Path:     buildLocalCacheDir(u.Host, u.Path, revision),
		URL:      u.String(),
		Revision: revision,
	}, nil
}

func buildLocalCacheDir(host, path, revision string) string {
	revision = url.PathEscape(revision)

	return filepath.Join(LocalCache, host, fmt.Sprintf("%s@%s", path, revision))
}
