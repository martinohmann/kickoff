// Package skeleton contains types that define the structure of a skeleton and
// its config. It also provides helpers to find, create and merge skeletons.
package skeleton

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/martinohmann/kickoff/internal/file"
	"github.com/martinohmann/kickoff/internal/kickoff"
)

// ErrDirNotFound indicates that a skeleton directory was not found.
var ErrDirNotFound = errors.New("skeleton dir not found")

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
