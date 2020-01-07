package repo

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/martinohmann/kickoff/pkg/skeleton"
)

// IsSkeletonDir returns true if dir is a skeleton dir. Skeleton dirs are
// detected by the fact that they contain a .kickoff.yaml file.
func isSkeletonDir(dir string) (bool, error) {
	configPath := filepath.Join(dir, skeleton.ConfigFile)

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
func findSkeletons(dir string) ([]*skeleton.Info, error) {
	skeletons := make([]*skeleton.Info, 0)

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
		if err != nil {
			return nil, err
		}

		if ok {
			abspath, err := filepath.Abs(path)
			if err != nil {
				return nil, err
			}

			skeletons = append(skeletons, &skeleton.Info{
				Name: info.Name(),
				Path: abspath,
			})
			// We do not stop here as we also want to find nested skeletons.
		}

		skels, err := findSkeletons(path)
		if err != nil {
			return nil, err
		}

		for _, s := range skels {
			skeletons = append(skeletons, &skeleton.Info{
				Name: filepath.Join(info.Name(), s.Name),
				Path: s.Path,
			})
		}
	}

	return skeletons, nil
}
