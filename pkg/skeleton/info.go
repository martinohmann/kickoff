package skeleton

import (
	"os"
	"path/filepath"
)

const (
	// ConfigFile is the name of the skeleton's config file.
	ConfigFile = ".kickoff.yaml"
)

// Info holds information about a skeleton.
type Info struct {
	Name string
	Path string
}

// Config loads the skeleton config for the info.
func (i *Info) Config() (*Config, error) {
	return LoadConfig(filepath.Join(i.Path, ConfigFile))
}

// Walk recursively walks all files and directories of the skeleton. This
// behaves exactly as filepath.Walk, except that it will ignore the skeleton's
// ConfigFile.
func (i *Info) Walk(walkFn filepath.WalkFunc) error {
	return filepath.Walk(i.Path, func(path string, info os.FileInfo, err error) error {
		if info.Name() == ConfigFile {
			// ignore skeleton config file
			return err
		}

		return walkFn(path, info, err)
	})
}
