package repository

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/skeleton"
	log "github.com/sirupsen/logrus"
)

// LocalRepository is a local skeleton repository. A local skeleton repository
// can be any directory on disk that contains a skeletons/ subdirectory.
type LocalRepository struct {
	ref kickoff.RepoRef
}

// NewLocalRepository creates a *LocalRepository from ref. Returns an error if
// resolving the absolute path to the skeleton repository fails.
func NewLocalRepository(ref kickoff.RepoRef) (*LocalRepository, error) {
	var err error

	ref.Path, err = filepath.Abs(ref.Path)
	if err != nil {
		return nil, err
	}

	return &LocalRepository{ref: ref}, nil
}

// GetSkeleton implements Repository.
func (r *LocalRepository) GetSkeleton(ctx context.Context, name string) (*skeleton.Info, error) {
	path := filepath.Join(r.ref.Path, "skeletons", name)

	ok, err := skeleton.IsSkeletonDir(path)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, SkeletonNotFoundError{name, r.ref.Name}
	}

	info := &skeleton.Info{
		Name: name,
		Path: path,
		Repo: &r.ref,
	}

	return info, nil
}

// ListSkeletons implements Repository.
func (r *LocalRepository) ListSkeletons(ctx context.Context) ([]*skeleton.Info, error) {
	infos, err := findSkeletons(&r.ref, filepath.Join(r.ref.Path, "skeletons"))
	if err != nil {
		return nil, fmt.Errorf("failed to list skeletons: %w", err)
	}

	return infos, nil
}

func findSkeletons(repoRef *kickoff.RepoRef, dir string) ([]*skeleton.Info, error) {
	skeletons := make([]*skeleton.Info, 0)

	fileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, info := range fileInfos {
		if !info.IsDir() {
			continue
		}

		path := filepath.Join(dir, info.Name())

		ok, err := skeleton.IsSkeletonDir(path)
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

			skeletons = append(skeletons, &skeleton.Info{
				Name: info.Name(),
				Path: abspath,
				Repo: repoRef,
			})
			continue
		}

		skels, err := findSkeletons(repoRef, path)
		if err != nil {
			return nil, err
		}

		for _, s := range skels {
			skeletons = append(skeletons, &skeleton.Info{
				Name: filepath.Join(info.Name(), s.Name),
				Path: s.Path,
				Repo: repoRef,
			})
		}
	}

	return skeletons, nil
}
