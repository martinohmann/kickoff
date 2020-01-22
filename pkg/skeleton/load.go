package skeleton

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/martinohmann/kickoff/pkg/template"
)

// File contains paths and other information about a skeleton file, e.g.
// whether it was inherited from a parent skeleton or not.
type File struct {
	// RelPath is the file path relative to root directory of the skeleton.
	RelPath string

	// AbsPath is the absolute path to the file on disk.
	AbsPath string

	// Inherited indicates whether the file was inherited from a parent
	// skeleton or not.
	Inherited bool

	// Info is the os.FileInfo for the file
	Info os.FileInfo
}

// Skeleton is the representation of a skeleton returned by Load() with all
// references to parent skeletons (if any) resolved.
type Skeleton struct {
	// Description is the skeleton description text obtained from the skeleton
	// config.
	Description string

	// Parent is a reference to the parent skeleton. Is nil if the skeleton has
	// no parent.
	Parent *Skeleton

	// Info is the skeleton info that was used to load the skeleton.
	Info *Info

	// The Files slice contains a merged and sorted list of file references
	// that includes all files from the skeleton and its parents (if any).
	Files []*File

	// Values are the template values from the skeleton's config merged with
	// those of it's parents (if any).
	Values template.Values

	fileMap map[string]*File
}

// WalkFiles walks all skeleton files using fn.
func (s *Skeleton) WalkFiles(fn func(file *File, err error) error) error {
	var err error

	for _, file := range s.Files {
		err = fn(file, err)
		if err != nil {
			return err
		}
	}

	return nil
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

	filePaths := make([]string, 0, len(s.fileMap))
	for path := range s.fileMap {
		filePaths = append(filePaths, path)
	}

	sort.Strings(filePaths)

	files := make([]*File, len(filePaths))
	for i, path := range filePaths {
		files[i] = s.fileMap[path]
	}

	s.Files = files

	return s, nil
}

func load(info *Info, visits map[Reference]struct{}) (*Skeleton, error) {
	config, err := info.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load skeleton config: %v", err)
	}

	s := &Skeleton{
		Description: config.Description,
		Values:      config.Values,
		Info:        info,

		fileMap: make(map[string]*File),
	}

	if ref := config.Parent; ref != nil {
		s.Parent, err = loadParent(info, ref, visits)
		if err != nil {
			return nil, err
		}

		for k, v := range s.Parent.fileMap {
			s.fileMap[k] = &File{
				AbsPath:   v.AbsPath,
				RelPath:   v.RelPath,
				Info:      v.Info,
				Inherited: true,
			}
		}

		err = mergeValues(s)
		if err != nil {
			return nil, err
		}
	}

	err = collectFiles(s)
	if err != nil {
		return nil, err
	}

	return s, nil
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

func mergeValues(s *Skeleton) error {
	values := template.Values{}

	err := values.Merge(s.Parent.Values)
	if err != nil {
		return err
	}

	err = values.Merge(s.Values)
	if err != nil {
		return err
	}

	s.Values = values

	return nil
}

func collectFiles(s *Skeleton) error {
	return s.Info.Walk(func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(s.Info.Path, path)
		if err != nil {
			return err
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		s.fileMap[relPath] = &File{
			RelPath: relPath,
			AbsPath: absPath,
			Info:    fi,
		}

		return nil
	})
}
