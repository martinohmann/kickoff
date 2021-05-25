package kickoff

import (
	"fmt"

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
	// Parameters holds the schema of parameters that can be set by the user
	// when creating a project from the skeleton.
	Parameters Parameters `json:"parameters,omitempty"`
}

// Validate implements the Validator interface.
func (c *SkeletonConfig) Validate() error {
	return c.Parameters.Validate()
}

// LoadSkeletonConfig loads the skeleton config from path and returns it.
func LoadSkeletonConfig(path string) (*SkeletonConfig, error) {
	var config SkeletonConfig

	if err := Load(path, &config); err != nil {
		return nil, fmt.Errorf("failed to load skeleton config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &config, nil
}
