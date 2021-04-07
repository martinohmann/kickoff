package kickoff

import (
	"github.com/martinohmann/kickoff/internal/template"
)

// SkeletonConfig describes the schema of a skeleton's .kickoff.yaml that is
// bundled together with the skeleton.
type SkeletonConfig struct {
	// Description holds a decription of the skeleton that can give some more
	// user-defined hints on the skeleton usage, e.g. interesting values to
	// tweak.
	Description string `json:"description,omitempty"`
	// Parent references an optional parent skeleton.
	Parent *ParentRef `json:"parent,omitempty"`
	// Values holds user-defined values available in .skel templates.
	Values template.Values `json:"values,omitempty"`
}

// ParentRef is a reference to a parent skeleton, possibly in a different
// repository.
type ParentRef struct {
	// SkeletonName holds the name of the parent skeleton. May include slashes
	// if the skeleton is organized in a subdirectory relative to the skeletons
	// root.
	SkeletonName string `json:"skeletonName"`
	// RepositoryURL can be a local path or are remote url to a skeleton
	// repository. If empty the parent is assumed to be in the same repository
	// as the child.
	RepositoryURL string `json:"repositoryURL,omitempty"`
}

// LoadSkeletonConfig loads the skeleton config from path and returns it.
func LoadSkeletonConfig(path string) (*SkeletonConfig, error) {
	var config SkeletonConfig

	err := Load(path, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
