package kickoff

import (
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
