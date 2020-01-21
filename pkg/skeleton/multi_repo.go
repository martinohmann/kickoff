package skeleton

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

var (
	// ErrNoRepositories is returned by NewMultiRepo if no repositories are
	// configured.
	ErrNoRepositories = errors.New("no skeleton repositories configured")
)

type multiRepo struct {
	repoURLs  map[string]string
	repoNames []string
}

// NewMultiRepo creates a new Repository which will lookup skeletons in the
// repositories provided in the repos map. The map maps an arbitrary repository
// name to an url.
func NewMultiRepo(repos map[string]string) (Repository, error) {
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

	r := &multiRepo{
		repoURLs:  repos,
		repoNames: repoNames,
	}

	return r, nil
}

func (r *multiRepo) SkeletonInfo(name string) (*Info, error) {
	repoName, name := splitName(name)
	if repoName == "" {
		return r.findSkeleton(name)
	}

	url, ok := r.repoURLs[repoName]
	if !ok {
		return nil, fmt.Errorf("no skeleton repository configured with name %q", repoName)
	}

	repo, err := openNamedRepository(repoName, url)
	if err != nil {
		return nil, err
	}

	return repo.SkeletonInfo(name)
}

func (r *multiRepo) LoadSkeleton(name string) (*Skeleton, error) {
	info, err := r.SkeletonInfo(name)
	if err != nil {
		return nil, err
	}

	return Load(info)
}

func (r *multiRepo) findSkeleton(name string) (*Info, error) {
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
		"skeleton %q found in multiple repositories: %s. explicitly provide <repo-name>:<skeleton-name> to select one",
		name,
		strings.Join(seenRepos, ", "),
	)
}

func (r *multiRepo) SkeletonInfos() ([]*Info, error) {
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

func (r *multiRepo) foreachRepository(fn func(repo Repository) error) error {
	for _, name := range r.repoNames {
		url := r.repoURLs[name]

		repo, err := openNamedRepository(name, url)
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

// splitName splits the name into repo name and skeleton name. Repo name may be
// empty if name is not explicitly prefixed with a repository.
func splitName(name string) (string, string) {
	parts := strings.SplitN(name, ":", 2)
	if len(parts) < 2 {
		return "", parts[0]
	}

	return parts[0], parts[1]
}
