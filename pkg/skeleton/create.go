package skeleton

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/apex/log"
)

// Create creates a new skeleton at path. The created skeleton contains an
// example .kickoff.yaml and example README.md.skel as starter. Returns an
// error if creating path or writing any of the files fails.
func Create(path string) error {
	log.WithField("path", path).Info("creating skeleton directory")

	err := os.MkdirAll(path, 0755)
	if err != nil {
		return fmt.Errorf("failed to create skeleton dir %q", err)
	}

	err = writeReadmeSkeleton(path)
	if err != nil {
		return fmt.Errorf("failed to write skeleton README: %v", err)
	}

	err = writeConfigFile(path)
	if err != nil {
		return fmt.Errorf("failed to write skeleton config: %v", err)
	}

	return nil
}

// CreateRepository creates a new skeleton repository at path and initializes
// it with a skeleton located in a subdir named skeletonName.
func CreateRepository(path, skeletonName string) error {
	log.WithField("path", path).Info("creating skeleton repository")

	err := os.MkdirAll(path, 0755)
	if err != nil {
		return fmt.Errorf("failed to create skeleton repository %q", err)
	}

	skeletonDir := filepath.Join(path, skeletonName)

	return Create(skeletonDir)
}

func writeReadmeSkeleton(dir string) error {
	readmeSkelPath := filepath.Join(dir, "README.md.skel")

	log.WithField("path", readmeSkelPath).Info("writing README.md.skel")

	return ioutil.WriteFile(readmeSkelPath, defaultReadmeSkeletonBytes, 0644)
}

func writeConfigFile(dir string) error {
	configPath := filepath.Join(dir, ConfigFileName)

	log.WithField("path", configPath).Infof("writing %s", ConfigFileName)

	return ioutil.WriteFile(configPath, defaultConfigBytes, 0644)
}
