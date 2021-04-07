package skeleton

import (
	"fmt"
	"path/filepath"

	"github.com/martinohmann/kickoff/internal/kickoff"
)

// Info holds information about the location of a skeleton.
type Info struct {
	Name string           `json:"name"`
	Path string           `json:"path"`
	Repo *kickoff.RepoRef `json:"repo"`
}

// String implements fmt.Stringer.
func (i *Info) String() string {
	if i.Repo == nil || i.Repo.Name == "" {
		return i.Name
	}

	return fmt.Sprintf("%s:%s", i.Repo.Name, i.Name)
}

// LoadConfig loads the skeleton config for the info.
func (i *Info) LoadConfig() (*kickoff.SkeletonConfig, error) {
	configPath := filepath.Join(i.Path, kickoff.SkeletonConfigFileName)

	return kickoff.LoadSkeletonConfig(configPath)
}
