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
	// Values holds user-defined values available in .skel templates.
	Values template.Values `json:"values,omitempty"`
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
