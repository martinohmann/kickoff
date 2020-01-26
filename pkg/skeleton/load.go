package skeleton

import (
	"fmt"
	"os"
	"path/filepath"
)

// Load loads a skeleton based on its *Info. It will recursively load all
// parent skeletons (if any) and merge all parent values and files into the
// resulting *Skeleton.
func Load(info *Info) (*Skeleton, error) {
	visits := make(map[Reference]struct{})

	s, err := load(info, visits)
	if err != nil {
		return nil, fmt.Errorf("failed to load skeleton: %v", err)
	}

	return s, nil
}

// load loads a skeleton and tracks all visited parents in a map. It will
// recursively load and merge all parents into the skeleton. Returns an error
// if a dependency cycle is detected while loading a parent.
func load(info *Info, visits map[Reference]struct{}) (*Skeleton, error) {
	config, err := info.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load skeleton config: %v", err)
	}

	files, err := collectFiles(info)
	if err != nil {
		return nil, err
	}

	s := &Skeleton{
		Description: config.Description,
		Values:      config.Values,
		Info:        info,
		Files:       files,
	}

	if config.Parent == nil {
		return s, nil
	}

	parent, err := loadParent(info, config.Parent, visits)
	if err != nil {
		return nil, err
	}

	return Merge(parent, s)
}

func loadParent(info *Info, ref *Reference, visits map[Reference]struct{}) (*Skeleton, error) {
	if _, ok := visits[*ref]; ok {
		return nil, fmt.Errorf("dependency cycle detected for parent: %#v", *ref)
	}

	repoURL := ref.RepositoryURL

	if len(repoURL) == 0 {
		// If no repository url is provided we assume that the parent resides
		// in the same repo as the child.
		repoURL = info.Repo.LocalPath()
	}

	repoInfo, err := ParseRepositoryURL(repoURL)
	if err != nil {
		return nil, err
	}

	// It is allowed to reference skeletons in the same repository
	// using relative URLs, so we have to account for that and prefix
	// the URL with the path of the child skeleton.
	if repoInfo.Local && !filepath.IsAbs(repoURL) {
		repoURL = filepath.Join(info.Path, repoURL)
	}

	repo, err := OpenRepository(repoURL)
	if err != nil {
		return nil, err
	}

	parent, err := repo.SkeletonInfo(ref.SkeletonName)
	if err != nil {
		return nil, err
	}

	visits[*ref] = struct{}{}

	return load(parent, visits)
}

func collectFiles(info *Info) ([]*File, error) {
	files := make([]*File, 0)

	err := info.Walk(func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(info.Path, path)
		if err != nil {
			return err
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		files = append(files, &File{
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
