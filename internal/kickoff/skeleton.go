package kickoff

import (
	"fmt"
	"path/filepath"

	"github.com/martinohmann/kickoff/internal/template"
)

// Skeleton is the representation of a skeleton loaded from a skeleton
// repository.
type Skeleton struct {
	// Description is the skeleton description text obtained from the skeleton
	// config.
	Description string `json:"description,omitempty"`
	// Ref holds the information about the location that was used to load the
	// skeleton. May be nil if the skeleton is the merged result of composing
	// multiple skeletons.
	Ref *SkeletonRef `json:"ref,omitempty"`
	// Parent points to the parent skeleton if there is one.
	Parent *Skeleton `json:"parent,omitempty"`
	// The Files slice contains a merged and sorted list of file references
	// that includes all files from the skeleton and its parents (if any).
	Files []*FileRef `json:"files,omitempty"`
	// Values are the template values from the skeleton's metadata merged with
	// those of it's parents (if any).
	Values template.Values `json:"values,omitempty"`
}

// String implements fmt.Stringer.
func (s *Skeleton) String() string {
	name := "<anonymous-skeleton>"
	if s.Ref != nil {
		name = s.Ref.String()
	}

	if s.Parent == nil {
		return name
	}

	return fmt.Sprintf("%s->%s", s.Parent, name)
}

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

// Validate implements the Validator interface.
func (r *SkeletonRef) Validate() error {
	if r.Name == "" {
		return newSkeletonRefError("Name must not be empty")
	}

	if r.Repo != nil {
		return r.Repo.Validate()
	}

	return nil
}

// LoadConfig loads the skeleton config for the info.
func (r *SkeletonRef) LoadConfig() (*SkeletonConfig, error) {
	configPath := filepath.Join(r.Path, SkeletonConfigFileName)

	return LoadSkeletonConfig(configPath)
}
