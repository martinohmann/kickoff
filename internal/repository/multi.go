package repository

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/martinohmann/kickoff/internal/skeleton"
)

// MultiRepository is a repository that aggregates multiple repositories and
// implements the Repository interface.
type MultiRepository struct {
	repoNames []string
	repoMap   map[string]Repository
}

// NewMultiRepository creates a *MultiRepository which aggregates the
// repositores from the repoURLMap. The repoURLMap is a mapping of repository
// name to its url. Returns an error if repoURLMap contains empty keys or if
// creating individual repositories fails, or if repoURLMap is empty.
func NewMultiRepository(repoURLMap map[string]string) (*MultiRepository, error) {
	if len(repoURLMap) == 0 {
		return nil, ErrNoRepositories
	}

	r := &MultiRepository{
		repoNames: make([]string, 0, len(repoURLMap)),
		repoMap:   make(map[string]Repository, len(repoURLMap)),
	}

	for name, url := range repoURLMap {
		if name == "" {
			return nil, fmt.Errorf("repository with url %s was configured with an empty name, please fix your config", url)
		}

		repo, err := NewNamed(name, url)
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

// GetSkeleton implements Repository.
//
// Attempts to find the named skeleton in any of the backing repositories and
// returns it. If the skeleton name is ambiguous, GetSkeleton will return an
// error. Repositories are traversed in the lexicographical order of their
// names. If name has the form `<repoName>:<skeletonName>`, the skeleton will
// be looked up in the repository that matched repoName. Returns
// SkeletonNotFoundError if the skeleton was not found in any of the configured
// repositories.
func (r *MultiRepository) GetSkeleton(ctx context.Context, name string) (*skeleton.Info, error) {
	repoName, skeletonName := splitName(name)
	if repoName == "" {
		return r.findSkeleton(ctx, skeletonName)
	}

	repo, ok := r.repoMap[repoName]
	if !ok {
		return nil, fmt.Errorf("no skeleton repository configured with name %q", repoName)
	}

	return repo.GetSkeleton(ctx, skeletonName)
}

// ListSkeletons implements Repository.
//
// Lists the skeletons of all configured repositories lexicographically sorted
// by repository name.
func (r *MultiRepository) ListSkeletons(ctx context.Context) ([]*skeleton.Info, error) {
	allSkeletons := make([]*skeleton.Info, 0)

	for _, repoName := range r.repoNames {
		repo := r.repoMap[repoName]

		skeletons, err := repo.ListSkeletons(ctx)
		if err != nil {
			return nil, err
		}

		allSkeletons = append(allSkeletons, skeletons...)
	}

	return allSkeletons, nil
}

func (r *MultiRepository) findSkeleton(ctx context.Context, name string) (*skeleton.Info, error) {
	candidates := make([]*skeleton.Info, 0)
	seenRepos := make([]string, 0)

	for _, repoName := range r.repoNames {
		repo := r.repoMap[repoName]

		skeleton, err := repo.GetSkeleton(ctx, name)
		if _, ok := err.(SkeletonNotFoundError); ok {
			// Ignore the error, we will return an error only if the skeleton
			// was not found in any of the repositories.
			continue
		} else if err != nil {
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
