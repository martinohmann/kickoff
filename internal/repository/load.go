package repository

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/martinohmann/kickoff/internal/kickoff"
)

// LoadSkeletons loads multiple skeletons from given repository. The passed in
// context is propagated to all operations that cross API boundaries (e.g. git
// operations) and can be used to enforce timeouts or cancel them. Returns an
// error if loading of any of the skeletons fails.
func LoadSkeletons(ctx context.Context, repo kickoff.Repository, names []string) ([]*kickoff.Skeleton, error) {
	skeletons := make([]*kickoff.Skeleton, len(names))

	for i, name := range names {
		skeleton, err := LoadSkeleton(ctx, repo, name)
		if err != nil {
			return nil, err
		}

		skeletons[i] = skeleton
	}

	return skeletons, nil
}

// LoadSkeleton loads the skeleton with name from given repository. The passed
// in context is propagated to all operations that cross API boundaries (e.g.
// git operations) and can be used to enforce timeouts or cancel them. Returns
// an error if loading the skeleton fails.
func LoadSkeleton(ctx context.Context, repo kickoff.Repository, name string) (*kickoff.Skeleton, error) {
	info, err := repo.GetSkeleton(ctx, name)
	if err != nil {
		return nil, err
	}

	visits := make(map[kickoff.ParentRef]struct{})

	s, err := loadSkeleton(ctx, info, visits)
	if err != nil {
		return nil, fmt.Errorf("failed to load skeleton: %w", err)
	}

	return s, nil
}

// loadSkeleton loads a skeleton and tracks all visited parents in a map. It will
// recursively load and merge all parents into the skeleton. Returns an error
// if a dependency cycle is detected while loading a parent.
func loadSkeleton(ctx context.Context, ref *kickoff.SkeletonRef, visits map[kickoff.ParentRef]struct{}) (*kickoff.Skeleton, error) {
	config, err := ref.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load skeleton config: %w", err)
	}

	files, err := collectFiles(ref)
	if err != nil {
		return nil, err
	}

	s := &kickoff.Skeleton{
		Description: config.Description,
		Values:      config.Values,
		Ref:         ref,
		Files:       files,
	}

	if config.Parent == nil {
		return s, nil
	}

	parent, err := loadParent(ctx, ref, *config.Parent, visits)
	if err != nil {
		return nil, err
	}

	return parent.Merge(s)
}

func loadParent(ctx context.Context, ref *kickoff.SkeletonRef, parentRef kickoff.ParentRef, visits map[kickoff.ParentRef]struct{}) (*kickoff.Skeleton, error) {
	if _, ok := visits[parentRef]; ok {
		return nil, DependencyCycleError{ParentRef: parentRef}
	}

	parentRepoRef, err := getParentRepoRef(ref.Repo, parentRef)
	if err != nil {
		return nil, err
	}

	repo, err := NewFromRef(*parentRepoRef)
	if err != nil {
		return nil, err
	}

	parent, err := repo.GetSkeleton(ctx, parentRef.SkeletonName)
	if err != nil {
		return nil, err
	}

	visits[parentRef] = struct{}{}

	return loadSkeleton(ctx, parent, visits)
}

func getParentRepoRef(repoRef *kickoff.RepoRef, parentRef kickoff.ParentRef) (*kickoff.RepoRef, error) {
	if parentRef.RepositoryURL == "" {
		return repoRef, nil
	}

	parentRepoRef, err := kickoff.ParseRepoRef(parentRef.RepositoryURL)
	if err != nil {
		return nil, err
	}

	if repoRef.IsRemote() && parentRepoRef.IsLocal() {
		return nil, fmt.Errorf("cannot reference skeleton from local path %q as parent in remote repository %q", parentRepoRef.Path, repoRef)
	}

	if parentRepoRef.IsLocal() && !filepath.IsAbs(parentRepoRef.Path) {
		// It is allowed to reference local parent skeletons using relative
		// paths, so we have to account for that and prefix the path with the
		// path of the child skeleton.
		localPath := repoRef.LocalPath()

		parentRepoRef.Path = filepath.Join(localPath, parentRepoRef.Path)
	}

	return parentRepoRef, nil
}

func collectFiles(ref *kickoff.SkeletonRef) ([]kickoff.File, error) {
	files := make([]kickoff.File, 0)

	err := filepath.Walk(ref.Path, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi.Name() == kickoff.SkeletonConfigFileName {
			// ignore skeleton config file
			return nil
		}

		relPath, err := filepath.Rel(ref.Path, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			// ignore skeleton dir itself
			return nil
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		files = append(files, &kickoff.FileRef{
			RelPath:  relPath,
			AbsPath:  absPath,
			FileMode: fi.Mode(),
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}
