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

var _ kickoff.Repository = (*localRepository)(nil)

// localRepository is a local skeleton repository. A local skeleton repository
// can be any directory on disk that contains a skeletons/ subdirectory.
type localRepository struct {
	ref kickoff.RepoRef
}

// newLocal creates a *localRepository from ref. Returns an error if
// resolving the absolute path to the skeleton repository fails.
func newLocal(ref kickoff.RepoRef) *localRepository {
	return &localRepository{ref: ref}
}

// GetSkeleton implements kickoff.Repository.
func (r *localRepository) GetSkeleton(ctx context.Context, name string) (*kickoff.SkeletonRef, error) {
	path := r.ref.SkeletonPath(name)

	ok, err := skeleton.IsSkeletonDir(path)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, SkeletonNotFoundError{name, r.ref.Name}
	}

	info := &kickoff.SkeletonRef{
		Name: name,
		Path: path,
		Repo: &r.ref,
	}

	return info, nil
}

// ListSkeletons implements kickoff.Repository.
func (r *localRepository) ListSkeletons(ctx context.Context) ([]*kickoff.SkeletonRef, error) {
	infos, err := findSkeletons(&r.ref, r.ref.SkeletonsPath())
	if err != nil {
		return nil, fmt.Errorf("failed to list skeletons: %w", err)
	}

	return infos, nil
}

func findSkeletons(repoRef *kickoff.RepoRef, dir string) ([]*kickoff.SkeletonRef, error) {
	skeletons := make([]*kickoff.SkeletonRef, 0)

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

			skeletons = append(skeletons, &kickoff.SkeletonRef{
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
			skeletons = append(skeletons, &kickoff.SkeletonRef{
				Name: filepath.Join(info.Name(), s.Name),
				Path: s.Path,
				Repo: repoRef,
			})
		}
	}

	return skeletons, nil
}
