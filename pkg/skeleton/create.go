package skeleton

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/martinohmann/kickoff/pkg/boilerplate"
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

	readmeSkelPath := filepath.Join(path, "README.md.skel")

	log.WithField("path", readmeSkelPath).Info("writing README.md.skel")

	err = ioutil.WriteFile(readmeSkelPath, boilerplate.DefaultReadmeBytes(), 0644)
	if err != nil {
		return fmt.Errorf("failed to write skeleton README: %v", err)
	}

	configPath := filepath.Join(path, ConfigFileName)

	log.WithField("path", configPath).Infof("writing %s", ConfigFileName)

	err = ioutil.WriteFile(configPath, boilerplate.DefaultSkeletonConfigBytes(), 0644)
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
