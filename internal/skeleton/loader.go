package skeleton

import (
	"fmt"
	"os"
	"path/filepath"
)

// Loader loads skeletons from a repository.
type Loader struct {
	Repo Repository
}

// NewLoader creates a new *Loader value which will use repo to load skeletons.
func NewLoader(repo Repository) *Loader {
	return &Loader{
		Repo: repo,
	}
}

// NewSingleRepositoryLoader creates a new *Loader value which can load
// skeletons from the single repository at url. Returns an error if opening the
// repository fails.
func NewSingleRepositoryLoader(url string) (*Loader, error) {
	repo, err := OpenRepository(url)
	if err != nil {
		return nil, err
	}

	return NewLoader(repo), nil
}

// NewRepositoryAggregateLoader creates a new *Loader value which can load
// skeletons from multiple repositories. The repos map must contain repo names
// as keys and repo urls as values. Returns an error if the passed in map of
// repositories is invalid.
func NewRepositoryAggregateLoader(repos map[string]string) (*Loader, error) {
	repo, err := NewRepositoryAggregate(repos)
	if err != nil {
		return nil, err
	}

	return NewLoader(repo), nil
}

// LoadSkeleton loads the skeleton with name from the repository. The
// returned skeleton already includes the recursively merged list of files
// and values from potential parents.
func (l *Loader) LoadSkeleton(name string) (*Skeleton, error) {
	info, err := l.Repo.SkeletonInfo(name)
	if err != nil {
		return nil, err
	}

	return Load(info)
}

// LoadSkeletons loads multiple skeletons from the repository. The returned
// skeletons already includes the recursively merged list of files and
// values from potential parents.
func (l *Loader) LoadSkeletons(names []string) ([]*Skeleton, error) {
	var err error

	skeletons := make([]*Skeleton, len(names))

	for i, name := range names {
		skeletons[i], err = l.LoadSkeleton(name)
		if err != nil {
			return nil, err
		}
	}

	return skeletons, nil
}

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

	err := filepath.Walk(info.Path, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi.Name() == ConfigFileName {
			// ignore skeleton config file
			return nil
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
