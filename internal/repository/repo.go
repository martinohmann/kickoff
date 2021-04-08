package repository

import "github.com/martinohmann/kickoff/internal/kickoff"

// New creates a new kickoff Repository for url. Returns an error if url is not
// a valid local path or remote url.
func New(url string) (kickoff.Repository, error) {
	return NewNamed("", url)
}

// NewNamed creates a new named Repository. The name is propagated into the
// repository info that is attached to every skeleton that is retrieved from
// it. Apart from that is behaves exactly like New.
func NewNamed(name, url string) (kickoff.Repository, error) {
	ref, err := kickoff.ParseRepoRef(url)
	if err != nil {
		return nil, err
	}

	ref.Name = name

	return NewFromRef(*ref)
}

// NewFromRef creates a new kickoff.Repository from a repository reference. Ref
// may reference a local or remote repository.
func NewFromRef(ref kickoff.RepoRef) (kickoff.Repository, error) {
	key := cacheKey{ref.Name, ref.String()}

	if repo, ok := repoCache.get(key); ok {
		return repo, nil
	}

	repo, err := newFromRef(ref)
	if err != nil {
		return nil, err
	}

	repoCache.set(key, repo)

	return repo, nil
}

func newFromRef(ref kickoff.RepoRef) (kickoff.Repository, error) {
	if ref.IsRemote() {
		return newRemote(ref)
	}

	return newLocal(ref)
}

// NewFromMap creates a kickoff.Repository which aggregates the
// repositores from the repoURLMap. The repoURLMap is a mapping of repository
// name to its url. Returns an error if repoURLMap contains empty keys or if
// creating individual repositories fails, or if repoURLMap is empty.
func NewFromMap(repoURLMap map[string]string) (kickoff.Repository, error) {
	return newMulti(repoURLMap)
}
