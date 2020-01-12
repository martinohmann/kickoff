package skeleton

import (
	"errors"
	"fmt"
	"strings"

	"github.com/martinohmann/kickoff/pkg/config"
)

var (
	ErrNoRepositories = errors.New("no skeleton repositories configured")
)

type multiRepo struct {
	repos config.Repositories
}

func NewMultiRepo(repos config.Repositories) (Repository, error) {
	if len(repos) == 0 {
		return nil, ErrNoRepositories
	}

	r := &multiRepo{
		repos: repos,
	}

	return r, nil
}

func (f *multiRepo) init() error { return nil }

func (f *multiRepo) Skeleton(name string) (*Info, error) {
	alias, name := splitAliasName(name)
	if alias == "" {
		return f.findSkeleton(name)
	}

	url, ok := f.repos[alias]
	if !ok {
		return nil, fmt.Errorf("no skeleton repository config found for alias %q", alias)
	}

	repo, err := OpenRepository(url)
	if err != nil {
		return nil, err
	}

	return repo.Skeleton(name)
}

func (f *multiRepo) findSkeleton(name string) (*Info, error) {
	candidates := make([]*Info, 0)
	aliases := make([]string, 0)

	err := f.foreachRepository(func(alias string, repo Repository) error {
		skeleton, err := repo.Skeleton(name)
		if err != nil {
			// Ignore the error, we will return an error only if the skeleton
			// was not found in any of the repositories.
			return nil
		}

		candidates = append(candidates, &Info{
			Name:      skeleton.Name,
			Path:      skeleton.Path,
			RepoAlias: alias,
		})

		aliases = append(aliases, alias)

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
		"skeleton %q found in multiple repositories: %v. explictly provide <alias>:<name> to select one",
		name,
		aliases,
	)
}

func (f *multiRepo) Skeletons() ([]*Info, error) {
	allSkeletons := make([]*Info, 0)

	err := f.foreachRepository(func(alias string, repo Repository) error {
		skeletons, err := repo.Skeletons()
		if err != nil {
			return err
		}

		for _, skeleton := range skeletons {
			allSkeletons = append(allSkeletons, &Info{
				Name:      skeleton.Name,
				Path:      skeleton.Path,
				RepoAlias: alias,
			})
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return allSkeletons, nil
}

func (f *multiRepo) foreachRepository(fn func(alias string, repo Repository) error) error {
	for alias, url := range f.repos {
		if alias == "" {
			return fmt.Errorf("repository with url %s has an empty alias, please fix your config", url)
		}

		repo, err := OpenRepository(url)
		if err != nil {
			return err
		}

		err = fn(alias, repo)
		if err != nil {
			return err
		}
	}

	return nil
}

func splitAliasName(fullname string) (string, string) {
	parts := strings.SplitN(fullname, ":", 2)
	if len(parts) < 2 {
		return "", parts[0]
	}

	return parts[0], parts[1]
}
