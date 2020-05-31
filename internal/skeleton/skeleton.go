// Package skeleton provides functionality to interact with local and remote
// skeleton repositories and to fetch the configuration values of any given
// skeleton.
package skeleton

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/martinohmann/kickoff/internal/file"
	"github.com/martinohmann/kickoff/internal/template"
)

// ErrDirNotFound indicates that a skeleton directory was not found.
var ErrDirNotFound = errors.New("skeleton dir not found")

// File contains paths and other information about a skeleton file, e.g.
// whether it was inherited from a parent skeleton or not.
type File struct {
	// RelPath is the file path relative to root directory of the skeleton.
	RelPath string `json:"relPath"`

	// AbsPath is the absolute path to the file on disk.
	AbsPath string `json:"absPath"`

	// Inherited indicates whether the file was inherited from a parent
	// skeleton or not.
	Inherited bool `json:"inherited"`

	// Info is the os.FileInfo for the file
	Info os.FileInfo `json:"-"`
}

// Path implements project.File.
func (f *File) Path() string {
	return f.RelPath
}

// Mode implements project.File.
func (f *File) Mode() os.FileMode {
	return f.Info.Mode()
}

// Reader implements project.File.
func (f *File) Reader() (io.Reader, error) {
	return os.Open(f.AbsPath)
}

// Skeleton is the representation of a skeleton returned by Load() with all
// references to parent skeletons (if any) resolved.
type Skeleton struct {
	// Description is the skeleton description text obtained from the skeleton
	// config.
	Description string `json:"description,omitempty"`

	// Parent is a reference to the parent skeleton. Is nil if the skeleton has
	// no parent.
	Parent *Skeleton `json:"parent,omitempty"`

	// Info is the skeleton info that was used to load the skeleton.
	Info *Info `json:"info"`

	// The Files slice contains a merged and sorted list of file references
	// that includes all files from the skeleton and its parents (if any).
	Files []*File `json:"files,omitempty"`

	// Values are the template values from the skeleton's config merged with
	// those of it's parents (if any).
	Values template.Values `json:"values,omitempty"`
}

// String implements fmt.Stringer.
func (s *Skeleton) String() string {
	if s.Parent == nil {
		return s.Info.String()
	}

	return fmt.Sprintf("%s->%s", s.Parent, s.Info)
}

// Merge merges multiple skeletons together and returns a new *Skeleton.
// The skeletons are merged left to right with template values, skeleton files
// and skeleton info of the rightmost skeleton taking preference over already
// existing values. Template values are recursively merged and may cause errors
// on type mismatch. The resulting *Skeleton will have the second-to-last
// skeleton set as its parent, the second-to-last will have the third-to-last
// as parent and so forth. The original skeletons are not altered. If only one
// skeleton is passed it will be returned as is without modification. Passing a
// slice with length of zero will return in an error.
func Merge(skeletons ...*Skeleton) (*Skeleton, error) {
	if len(skeletons) == 0 {
		return nil, errors.New("cannot merge empty list of skeletons")
	}

	if len(skeletons) == 1 {
		return skeletons[0], nil
	}

	lhs := skeletons[0]

	var err error

	for _, rhs := range skeletons[1:] {
		lhs, err = merge(lhs, rhs)
		if err != nil {
			return nil, err
		}
	}

	return lhs, nil
}

func merge(lhs, rhs *Skeleton) (*Skeleton, error) {
	values, err := template.MergeValues(lhs.Values, rhs.Values)
	if err != nil {
		return nil, fmt.Errorf("failed to merge skeleton %s and %s: %v", lhs.Info, rhs.Info, err)
	}

	s := &Skeleton{
		Values:      values,
		Files:       mergeFiles(lhs.Files, rhs.Files),
		Description: rhs.Description,
		Parent:      lhs,
		Info:        rhs.Info,
	}

	return s, nil
}

func mergeFiles(lhs, rhs []*File) []*File {
	fileMap := make(map[string]*File)

	for _, f := range lhs {
		fileMap[f.RelPath] = &File{
			AbsPath:   f.AbsPath,
			RelPath:   f.RelPath,
			Info:      f.Info,
			Inherited: true,
		}
	}

	for _, f := range rhs {
		fileMap[f.RelPath] = f
	}

	filePaths := make([]string, 0, len(fileMap))
	for path := range fileMap {
		filePaths = append(filePaths, path)
	}

	sort.Strings(filePaths)

	files := make([]*File, len(filePaths))
	for i, path := range filePaths {
		files[i] = fileMap[path]
	}

	return files
}

// FindSkeletonDir finds the skeleton dir path resides in. It walk up the
// filesystem tree and checks for each parent if it is a skeleton dir and
// returns its path if the check was successful. Returns ErrDirNotFound if path
// does not reside within a skeleton dir or one of its subdirectories. Returns
// any other errors that may occur while traversing the filesystem tree.
func FindSkeletonDir(path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	ok, err := file.IsDirectory(path)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	if !ok {
		path = filepath.Dir(path)
	}

	for len(path) > 1 {
		ok, err := IsSkeletonDir(path)
		if err != nil {
			return "", err
		}

		if ok {
			return path, nil
		}

		path = filepath.Dir(path)
	}

	return "", ErrDirNotFound
}

// IsInsideSkeletonDir returns true if path is inside a skeleton dir. If path
// is a skeleton dir itself, this will return false.
func IsInsideSkeletonDir(path string) (bool, error) {
	dir, err := FindSkeletonDir(path)
	if err == ErrDirNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}

	path, err = filepath.Abs(path)
	if err != nil {
		return false, err
	}

	return dir != path, nil
}

// IsSkeletonDir returns true if dir is a skeleton dir. Skeleton dirs are
// detected by the fact that they contain a .kickoff.yaml file.
func IsSkeletonDir(dir string) (bool, error) {
	configPath := filepath.Join(dir, ConfigFileName)

	info, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return !info.IsDir(), nil
}
