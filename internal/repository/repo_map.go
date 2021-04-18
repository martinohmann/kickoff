package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/martinohmann/kickoff/internal/kickoff"
)

// OpenMap opens multiple repositories and returns a kickoff.Repository which
// aggregates the repositores from the repoURLMap. The repoURLMap is a mapping
// of repository name to its url. Returns an error if repoURLMap contains empty
// keys or if creating individual repositories fails, or if repoURLMap is
// empty.
func OpenMap(ctx context.Context, repoURLMap map[string]string, opts *Options) (kickoff.Repository, error) {
	return newRepositoryMap(ctx, repoURLMap, opts)
}

// repositoryMap is a repository that aggregates multiple repositories and
// implements the kickoff.Repository interface.
type repositoryMap struct {
	repoNames []string
	repoMap   map[string]kickoff.Repository
}

func newRepositoryMap(ctx context.Context, repoURLMap map[string]string, opts *Options) (*repositoryMap, error) {
	if len(repoURLMap) == 0 {
		return nil, ErrNoRepositories
	}

	r := &repositoryMap{
		repoNames: make([]string, 0, len(repoURLMap)),
		repoMap:   make(map[string]kickoff.Repository, len(repoURLMap)),
	}

	for name, url := range repoURLMap {
		if name == "" {
			return nil, fmt.Errorf("repository with url %s was configured with an empty name, please fix your config", url)
		}

		repo, err := openNamed(ctx, name, url, opts)
		if err != nil {
			return nil, err
		}

		r.repoNames = append(r.repoNames, name)
		r.repoMap[name] = repo
	}

	// Sort repoNames for stable iteration order.
	sort.Strings(r.repoNames)

	return r, nil
}

// GetSkeleton implements kickoff.Repository.
//
// Attempts to find the named skeleton in any of the backing repositories and
// returns it. If the skeleton name is ambiguous, GetSkeleton will return an
// error. Repositories are traversed in the lexicographical order of their
// names. If name has the form `<repoName>:<skeletonName>`, the skeleton will
// be looked up in the repository that matched repoName. Returns
// SkeletonNotFoundError if the skeleton was not found in any of the configured
// repositories.
func (r *repositoryMap) GetSkeleton(name string) (*kickoff.SkeletonRef, error) {
	repoName, skeletonName := splitName(name)
	if repoName == "" {
		return r.findSkeleton(skeletonName)
	}

	repo, ok := r.repoMap[repoName]
	if !ok {
		return nil, fmt.Errorf("no skeleton repository configured with name %q", repoName)
	}

	return repo.GetSkeleton(skeletonName)
}

// ListSkeletons implements kickoff.Repository.
//
// Lists the skeletons of all configured repositories lexicographically sorted
// by repository name.
func (r *repositoryMap) ListSkeletons() ([]*kickoff.SkeletonRef, error) {
	allSkeletons := make([]*kickoff.SkeletonRef, 0)

	for _, repoName := range r.repoNames {
		repo := r.repoMap[repoName]

		skeletons, err := repo.ListSkeletons()
		if err != nil {
			return nil, err
		}

		allSkeletons = append(allSkeletons, skeletons...)
	}

	return allSkeletons, nil
}

func (r *repositoryMap) LoadSkeleton(name string) (*kickoff.Skeleton, error) {
	return loadSkeleton(r, name)
}

func (r *repositoryMap) CreateSkeleton(name string) (*kickoff.SkeletonRef, error) {
	var repo kickoff.Repository

	repoName, skeletonName := splitName(name)
	if repoName != "" {
		var ok bool
		repo, ok = r.repoMap[repoName]
		if !ok {
			return nil, fmt.Errorf("no skeleton repository configured with name %q", repoName)
		}
	} else if len(r.repoNames) == 1 {
		repo = r.repoMap[r.repoNames[0]]
	} else {
		return nil, fmt.Errorf(
			"ambiguous skeleton name %q: explicitly provide <repo-name>:%s to select a repository",
			name,
			name,
		)
	}

	return repo.CreateSkeleton(skeletonName)
}

func (r *repositoryMap) findSkeleton(name string) (*kickoff.SkeletonRef, error) {
	candidates := make([]*kickoff.SkeletonRef, 0)
	seenRepos := make([]string, 0)

	for _, repoName := range r.repoNames {
		repo := r.repoMap[repoName]

		skeleton, err := repo.GetSkeleton(name)
		if err != nil {
			var notFoundErr SkeletonNotFoundError
			if errors.As(err, &notFoundErr) {
				// Ignore the error, we will return an error only if the skeleton
				// was not found in any of the repositories.
				continue
			}

			return nil, err
		}

		candidates = append(candidates, skeleton)
		seenRepos = append(seenRepos, skeleton.Repo.Name)
	}

	switch len(candidates) {
	case 0:
		return nil, SkeletonNotFoundError{Name: name}
	case 1:
		return candidates[0], nil
	default:
		return nil, fmt.Errorf(
			"skeleton %q found in multiple repositories: %s. explicitly provide <repo-name>:%s to select one",
			name,
			strings.Join(seenRepos, ", "),
			name,
		)
	}
}

// splitName splits the name into repo name and skeleton name. Repo name may be
// empty if name is not explicitly prefixed with a repository.
func splitName(name string) (string, string) {
	parts := strings.SplitN(name, ":", 2)
	if len(parts) < 2 {
		return "", parts[0]
	}

	return parts[0], parts[1]
}
