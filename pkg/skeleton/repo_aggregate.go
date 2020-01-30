package skeleton

import (
	"fmt"
	"sort"
	"strings"
)

type repositoryAggregate struct {
	repoURLs  map[string]string
	repoNames []string
}

// NewRepositoryAggregate creates a new Repository which will lookup skeletons in the
// repositories provided in the repos map. The map maps an arbitrary repository
// name to an url.
func NewRepositoryAggregate(repos map[string]string) (Repository, error) {
	if len(repos) == 0 {
		return nil, ErrNoRepositories
	}

	repoNames := make([]string, 0, len(repos))
	for name, url := range repos {
		if name == "" {
			return nil, fmt.Errorf("repository with url %s was configured with an empty name, please fix your config", url)
		}

		repoNames = append(repoNames, name)
	}

	sort.Strings(repoNames)

	r := &repositoryAggregate{
		repoURLs:  repos,
		repoNames: repoNames,
	}

	return r, nil
}

// SkeletonInfo implements Repository.
func (r *repositoryAggregate) SkeletonInfo(name string) (*Info, error) {
	repoName, name := splitName(name)
	if repoName == "" {
		return r.findSkeleton(name)
	}

	repo, err := r.openRepository(repoName)
	if err != nil {
		return nil, err
	}

	return repo.SkeletonInfo(name)
}

// SkeletonInfos implements Repository.
func (r *repositoryAggregate) SkeletonInfos() ([]*Info, error) {
	allSkeletons := make([]*Info, 0)

	err := r.foreachRepository(func(repo Repository) error {
		skeletons, err := repo.SkeletonInfos()
		if err != nil {
			return err
		}

		allSkeletons = append(allSkeletons, skeletons...)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return allSkeletons, nil
}

func (r *repositoryAggregate) findSkeleton(name string) (*Info, error) {
	candidates := make([]*Info, 0)
	seenRepos := make([]string, 0)

	err := r.foreachRepository(func(repo Repository) error {
		skeleton, err := repo.SkeletonInfo(name)
		if err != nil {
			// Ignore the error, we will return an error only if the skeleton
			// was not found in any of the repositories.
			return nil
		}

		candidates = append(candidates, skeleton)

		seenRepos = append(seenRepos, skeleton.Repo.Name)

		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(candidates) == 1 {
		return candidates[0], nil
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("skeleton %q not found", name)
	}

	return nil, fmt.Errorf(
		"skeleton %q found in multiple repositories: %s. explicitly provide <repo-name>:%s to select one",
		name,
		strings.Join(seenRepos, ", "),
		name,
	)
}

func (r *repositoryAggregate) foreachRepository(fn func(repo Repository) error) error {
	for _, name := range r.repoNames {
		repo, err := r.openRepository(name)
		if err != nil {
			return err
		}

		err = fn(repo)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *repositoryAggregate) openRepository(name string) (Repository, error) {
	url, ok := r.repoURLs[name]
	if !ok {
		return nil, fmt.Errorf("no skeleton repository configured with name %q", name)
	}

	return openNamedRepository(name, url)
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
