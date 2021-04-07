package kickoff

import (
	"fmt"
	"path/filepath"
)

// SkeletonRef holds information about the location of a skeleton.
type SkeletonRef struct {
	// Name of the skeleton. May include slashes if the skeleton is organized
	// in a subdirectory relative to the skeletons root.
	Name string `json:"name"`
	// Path contains the local path to the skeleton.
	Path string `json:"path"`
	// Repository references the repository where the skeleton can be found.
	Repo *RepoRef `json:"repo"`
}

// String implements fmt.Stringer.
func (r *SkeletonRef) String() string {
	if r.Repo == nil || r.Repo.Name == "" {
		return r.Name
	}

	return fmt.Sprintf("%s:%s", r.Repo.Name, r.Name)
}

// LoadConfig loads the skeleton config for the info.
func (r *SkeletonRef) LoadConfig() (*SkeletonConfig, error) {
	configPath := filepath.Join(r.Path, SkeletonConfigFileName)

	return LoadSkeletonConfig(configPath)
}
