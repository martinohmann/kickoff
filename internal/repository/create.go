package repository

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/martinohmann/kickoff/internal/skeleton"
)

// Create creates a new skeleton repository at path and initializes
// it with a skeleton located in a subdir named skeletonName.
func Create(path, skeletonName string) error {
	skeletonsDir := filepath.Join(path, "skeletons")

	log.WithField("path", path).Info("creating skeleton repository")

	err := os.MkdirAll(skeletonsDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create skeleton repository %q", err)
	}

	skeletonDir := filepath.Join(skeletonsDir, skeletonName)

	return skeleton.Create(skeletonDir)
}
