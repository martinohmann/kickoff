package repository

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/martinohmann/kickoff/internal/skeleton"
)

// LoadSkeletons loads multiple skeletons from given repository. The passed in
// context is propagated to all operations that cross API boundaries (e.g. git
// operations) and can be used to enforce timeouts or cancel them. Returns an
// error if loading of any of the skeletons fails.
func LoadSkeletons(ctx context.Context, repo Repository, names []string) ([]*skeleton.Skeleton, error) {
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
func LoadSkeleton(ctx context.Context, repo Repository, name string) (*skeleton.Skeleton, error) {
	info, err := repo.GetSkeleton(ctx, name)
	if err != nil {
		return nil, err
	}

	visits := make(map[skeleton.Reference]struct{})

	s, err := loadSkeleton(ctx, info, visits)
	if err != nil {
		return nil, fmt.Errorf("failed to load skeleton: %w", err)
	}

	return s, nil
}

// loadSkeleton loads a skeleton and tracks all visited parents in a map. It will
// recursively load and merge all parents into the skeleton. Returns an error
// if a dependency cycle is detected while loading a parent.
func loadSkeleton(ctx context.Context, info *skeleton.Info, visits map[skeleton.Reference]struct{}) (*skeleton.Skeleton, error) {
	config, err := info.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load skeleton config: %w", err)
	}

	files, err := collectFiles(info)
	if err != nil {
		return nil, err
	}

	s := &skeleton.Skeleton{
		Description: config.Description,
		Values:      config.Values,
		Info:        info,
		Files:       files,
	}

	if config.Parent == nil {
		return s, nil
	}

	parent, err := loadParent(ctx, info, *config.Parent, visits)
	if err != nil {
		return nil, err
	}

	return skeleton.Merge(parent, s)
}

func loadParent(ctx context.Context, info *skeleton.Info, ref skeleton.Reference, visits map[skeleton.Reference]struct{}) (*skeleton.Skeleton, error) {
	if _, ok := visits[ref]; ok {
		return nil, DependencyCycleError{ParentRef: ref}
	}

	repoURL := ref.RepositoryURL
	repoName := ""

	if len(repoURL) == 0 {
		// If no repository url is provided we assume that the parent resides
		// in the same repo as the child.
		repoURL = info.Repo.Path
		repoName = info.Repo.Name
	}

	repoInfo, err := ParseURL(repoURL)
	if err != nil {
		return nil, err
	}

	// It is allowed to reference skeletons in the same repository
	// using relative URLs, so we have to account for that and prefix
	// the URL with the path of the child skeleton.
	if !repoInfo.IsRemote() && !filepath.IsAbs(repoURL) {
		repoURL = filepath.Join(info.Path, repoURL)
		repoName = info.Repo.Name
	}

	repo, err := NewNamed(repoName, repoURL)
	if err != nil {
		return nil, err
	}

	parent, err := repo.GetSkeleton(ctx, ref.SkeletonName)
	if err != nil {
		return nil, err
	}

	visits[ref] = struct{}{}

	return loadSkeleton(ctx, parent, visits)
}

func collectFiles(info *skeleton.Info) ([]*skeleton.File, error) {
	files := make([]*skeleton.File, 0)

	err := filepath.Walk(info.Path, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi.Name() == skeleton.ConfigFileName {
			// ignore skeleton config file
			return nil
		}

		relPath, err := filepath.Rel(info.Path, path)
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

		files = append(files, &skeleton.File{
			RelPath: relPath,
			AbsPath: absPath,
			Info:    fi,
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}
