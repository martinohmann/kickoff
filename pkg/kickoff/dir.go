package kickoff

import (
	"os"
	"path/filepath"

	"github.com/martinohmann/kickoff/pkg/config"
)

// IsSkeletonDir returns true if dir is a skeleton dir. Skeleton dirs are
// detected by the fact that they contain a .kickoff.yaml file.
func IsSkeletonDir(dir string) bool {
	configPath := filepath.Join(dir, config.SkeletonConfigFile)

	_, err := os.Stat(configPath)
	if err != nil {
		return false
	}

	return true
}
