package skeleton

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

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

	return writeFiles(path)
}

// CreateRepository creates a new skeleton repository at path and initializes
// it with a skeleton located in a subdir named skeletonName.
func CreateRepository(path, skeletonName string) error {
	skeletonsDir := filepath.Join(path, "skeletons")

	log.WithField("path", path).Info("creating skeleton repository")

	err := os.MkdirAll(skeletonsDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create skeleton repository %q", err)
	}

	skeletonDir := filepath.Join(skeletonsDir, skeletonName)

	return Create(skeletonDir)
}

func writeFiles(dir string) error {
	filenames := make([]string, 0, len(fileTemplates))
	for filename := range fileTemplates {
		filenames = append(filenames, filename)
	}

	sort.Strings(filenames)

	for _, filename := range filenames {
		path := filepath.Join(dir, filename)
		contents := fileTemplates[filename]

		log.WithField("path", path).Infof("writing %s", filename)

		err := ioutil.WriteFile(path, []byte(contents), 0644)
		if err != nil {
			return fmt.Errorf("failed to write skeleton file: %v", err)
		}
	}

	return nil
}
