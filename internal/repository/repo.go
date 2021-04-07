package repository

import (
	"github.com/kirsle/configdir"
	"github.com/martinohmann/kickoff/internal/kickoff"
)

// LocalCache holds the path to the local repository cache. This is platform
// specific.
var LocalCache = configdir.LocalCache("kickoff", "repositories")

// New creates a new Repository for url. Returns an error if url is not a valid
// local path or remote url.
func New(url string) (kickoff.Repository, error) {
	return NewNamed("", url)
}

// NewNamed creates a new named Repository. The name is propagated into the
// repository info that is attached to every skeleton that is retrieved from
// it. Apart from that is behaves exactly like New.
func NewNamed(name, url string) (repo kickoff.Repository, err error) {
	key := cacheKey{name, url}

	if repo, ok := repoCache.get(key); ok {
		return repo, nil
	}

	ref, err := kickoff.ParseRepoRef(url)
	if err != nil {
		return nil, err
	}

	ref.Name = name

	if ref.IsRemote() {
		repo, err = NewRemoteRepository(*ref)
	} else {
		repo, err = NewLocalRepository(*ref)
	}

	if err != nil {
		return nil, err
	}

	repoCache.set(key, repo)

	return repo, nil
}
