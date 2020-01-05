package kickoff

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/martinohmann/kickoff/pkg/config"
)

// IsSkeletonDir returns true if dir is a skeleton dir. Skeleton dirs are
// detected by the fact that they contain a .kickoff.yaml file.
func IsSkeletonDir(dir string) (bool, error) {
	configPath := filepath.Join(dir, config.SkeletonConfigFile)

	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

// FindSkeletons recursively finds all skeletons in dir. The resulting string
// slice contains paths relative to dir. Returns any error that may occur while
// traversing dir.
func FindSkeletons(dir string) ([]string, error) {
	skeletons := make([]string, 0)

	fileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, info := range fileInfos {
		if !info.IsDir() {
			continue
		}

		path := filepath.Join(dir, info.Name())

		ok, err := IsSkeletonDir(path)
		if err != nil {
			return nil, err
		}

		if ok {
			skeletons = append(skeletons, info.Name())
			// We do not stop here as we also want to find nested skeletons.
		}

		skels, err := FindSkeletons(path)
		if err != nil {
			return nil, err
		}

		for _, s := range skels {
			skeletons = append(skeletons, filepath.Join(info.Name(), s))
		}
	}

	return skeletons, nil
}
