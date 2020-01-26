// Package skeleton provides functionality to interact with local and remote
// skeleton repositories and to fetch the configuration values of any given
// skeleton.
package skeleton

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/martinohmann/kickoff/pkg/file"
	"github.com/martinohmann/kickoff/pkg/template"
)

var (
	// ErrDirNotFound indicates that a skeleton directory was not found.
	ErrDirNotFound = errors.New("skeleton dir not found")
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
}

// String implements fmt.Stringer.
func (s *Skeleton) String() string {
	if s.Parent == nil {
		return s.Info.String()
	}

	return fmt.Sprintf("%s->%s", s.Parent, s.Info)
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
		ok, err := isSkeletonDir(path)
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
func isSkeletonDir(dir string) (bool, error) {
	configPath := filepath.Join(dir, ConfigFileName)

	info, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return !info.IsDir(), nil
}

// findSkeletons recursively finds all skeletons in dir. Returns any error that
// may occur while traversing dir.
func findSkeletons(repo *RepositoryInfo, dir string) ([]*Info, error) {
	skeletons := make([]*Info, 0)

	fileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, info := range fileInfos {
		if !info.IsDir() {
			continue
		}

		path := filepath.Join(dir, info.Name())

		ok, err := isSkeletonDir(path)
		if os.IsPermission(err) {
			log.Warnf("permission error, skipping dir: %v", err)
			continue
		}

		if err != nil {
			return nil, err
		}

		if ok {
			abspath, err := filepath.Abs(path)
			if err != nil {
				return nil, err
			}

			skeletons = append(skeletons, &Info{
				Name: info.Name(),
				Path: abspath,
				Repo: repo,
			})
			continue
		}

		skels, err := findSkeletons(repo, path)
		if err != nil {
			return nil, err
		}

		for _, s := range skels {
			skeletons = append(skeletons, &Info{
				Name: filepath.Join(info.Name(), s.Name),
				Path: s.Path,
				Repo: repo,
			})
		}
	}

	return skeletons, nil
}
