package repository

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/skeleton"
)

// LoadSkeletons loads multiple skeletons from given repository. The passed in
// context is propagated to all operations that cross API boundaries (e.g. git
// operations) and can be used to enforce timeouts or cancel them. Returns an
// error if loading of any of the skeletons fails.
func LoadSkeletons(ctx context.Context, repo kickoff.Repository, names []string) ([]*skeleton.Skeleton, error) {
	skeletons := make([]*skeleton.Skeleton, len(names))

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
func LoadSkeleton(ctx context.Context, repo kickoff.Repository, name string) (*skeleton.Skeleton, error) {
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
func loadSkeleton(ctx context.Context, ref *kickoff.SkeletonRef, visits map[kickoff.ParentRef]struct{}) (*skeleton.Skeleton, error) {
	config, err := ref.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load skeleton config: %w", err)
	}

	files, err := collectFiles(ref)
	if err != nil {
		return nil, err
	}

	s := &skeleton.Skeleton{
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

	return skeleton.Merge(parent, s)
}

func loadParent(ctx context.Context, ref *kickoff.SkeletonRef, parentRef kickoff.ParentRef, visits map[kickoff.ParentRef]struct{}) (*skeleton.Skeleton, error) {
	if _, ok := visits[parentRef]; ok {
		return nil, DependencyCycleError{ParentRef: parentRef}
	}

	repoURL := parentRef.RepositoryURL
	repoName := ""

	if len(repoURL) == 0 {
		// If no repository url is provided we assume that the parent resides
		// in the same repo as the child.
		repoURL = ref.Repo.Path
		repoName = ref.Repo.Name
	}

	repoRef, err := kickoff.ParseRepoRef(repoURL)
	if err != nil {
		return nil, err
	}

	// It is allowed to reference skeletons in the same repository
	// using relative URLs, so we have to account for that and prefix
	// the URL with the path of the child skeleton.
	if !repoRef.IsRemote() && !filepath.IsAbs(repoURL) {
		repoURL = filepath.Join(ref.Path, repoURL)
		repoName = ref.Repo.Name
	}

	repo, err := NewNamed(repoName, repoURL)
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

func collectFiles(ref *kickoff.SkeletonRef) ([]*kickoff.FileRef, error) {
	files := make([]*kickoff.FileRef, 0)

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
