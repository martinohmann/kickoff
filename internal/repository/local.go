package repository

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/martinohmann/kickoff/internal/skeleton"
)

// LocalRepository is a local skeleton repository. A local skeleton repository
// can be any directory on disk that contains a skeletons/ subdirectory.
type LocalRepository struct {
	info skeleton.RepoInfo
}

// NewLocalRepository creates a *LocalRepository from info. Returns an error if
// resolving the absolute path to the skeleton repository fails.
func NewLocalRepository(info skeleton.RepoInfo) (*LocalRepository, error) {
	var err error

	info.Path, err = filepath.Abs(info.Path)
	if err != nil {
		return nil, err
	}

	return &LocalRepository{info: info}, nil
}

// GetSkeleton implements Repository.
func (r *LocalRepository) GetSkeleton(ctx context.Context, name string) (*skeleton.Info, error) {
	path := filepath.Join(r.info.Path, "skeletons", name)

	ok, err := skeleton.IsSkeletonDir(path)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, SkeletonNotFoundError{name, r.info.Name}
	}

	info := &skeleton.Info{
		Name: name,
		Path: path,
		Repo: &r.info,
	}

	return info, nil
}

// ListSkeletons implements Repository.
func (r *LocalRepository) ListSkeletons(ctx context.Context) ([]*skeleton.Info, error) {
	infos, err := r.info.FindSkeletons()
	if err != nil {
		return nil, fmt.Errorf("failed to list skeletons: %w", err)
	}

	return infos, nil
}
