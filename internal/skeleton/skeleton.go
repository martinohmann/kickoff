// Package skeleton contains types that define the structure of a skeleton and
// its config. It also provides helpers to find, create and merge skeletons.
package skeleton

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/martinohmann/kickoff/internal/file"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/template"
)

var (
	// ErrDirNotFound indicates that a skeleton directory was not found.
	ErrDirNotFound = errors.New("skeleton dir not found")

	// ErrMergeEmpty is returned by Merge if no skeletons were passed.
	ErrMergeEmpty = errors.New("cannot merge empty list of skeletons")
)

// Merge merges multiple skeletons together and returns a new *kickoff.Skeleton.
// The skeletons are merged left to right with template values, skeleton files
// and skeleton info of the rightmost skeleton taking preference over already
// existing values. Template values are recursively merged and may cause errors
// on type mismatch. The resulting *kickoff.Skeleton will have the second-to-last
// skeleton set as its parent, the second-to-last will have the third-to-last
// as parent and so forth. The original skeletons are not altered. If only one
// skeleton is passed it will be returned as is without modification. Passing a
// slice with length of zero will return in an error.
func Merge(skeletons ...*kickoff.Skeleton) (*kickoff.Skeleton, error) {
	if len(skeletons) == 0 {
		return nil, ErrMergeEmpty
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

func merge(lhs, rhs *kickoff.Skeleton) (*kickoff.Skeleton, error) {
	values, err := template.MergeValues(lhs.Values, rhs.Values)
	if err != nil {
		return nil, fmt.Errorf("failed to merge skeleton %s and %s: %w", lhs.Ref, rhs.Ref, err)
	}

	s := &kickoff.Skeleton{
		Values:      values,
		Files:       mergeFiles(lhs.Files, rhs.Files),
		Description: rhs.Description,
		Parent:      lhs,
		Ref:         rhs.Ref,
	}

	return s, nil
}

func mergeFiles(lhs, rhs []kickoff.File) []kickoff.File {
	fileMap := make(map[string]kickoff.File)

	for _, f := range lhs {
		fileMap[f.Path()] = f
	}

	for _, f := range rhs {
		fileMap[f.Path()] = f
	}

	filePaths := make([]string, 0, len(fileMap))
	for path := range fileMap {
		filePaths = append(filePaths, path)
	}

	sort.Strings(filePaths)

	files := make([]kickoff.File, len(filePaths))
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
	if errors.Is(err, ErrDirNotFound) {
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
	configPath := filepath.Join(dir, kickoff.SkeletonConfigFileName)

	info, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return !info.IsDir(), nil
}
