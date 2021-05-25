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
	// The Files slice contains a sorted list of files that are present in the
	// skeleton.
	Files []*BufferedFile `json:"files,omitempty"`
	// Values are the template values from the skeleton's metadata.
	Values template.Values `json:"values,omitempty"`
	// Parameters
	// @TODO(mohmann): add docs
	Parameters Parameters `json:"parameters,omitempty"`
}

// String implements fmt.Stringer.
func (s *Skeleton) String() string {
	name := "<anonymous-skeleton>"
	if s.Ref != nil {
		name = s.Ref.String()
	}

	return name
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

// MergeSkeletons merges multiple skeletons together and returns a new
// *Skeleton. The skeletons are merged left to right with template values,
// skeleton files and skeleton info of the rightmost skeleton taking preference
// over already existing values. Template values are recursively merged and may
// cause errors on type mismatch. The original skeletons are not altered. If
// only one skeleton is passed it will be returned as is without modification.
// Passing a slice with length of zero will return in an error.
func MergeSkeletons(skeletons ...*Skeleton) (*Skeleton, error) {
	if len(skeletons) == 0 {
		return nil, ErrMergeEmpty
	}

	if len(skeletons) == 1 {
		return skeletons[0], nil
	}

	s := skeletons[0]

	var err error

	for _, other := range skeletons[1:] {
		s, err = s.Merge(other)
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

// Merge merges two skeletons. The skeletons are merged left to right with
// template values, skeleton files and skeleton ref of the rightmost skeleton
// taking preference over already existing values. Template values are
// recursively merged and may cause errors on type mismatch. The original
// skeletons are not altered.
func (s *Skeleton) Merge(other *Skeleton) (*Skeleton, error) {
	values, err := template.MergeValues(s.Values, other.Values)
	if err != nil {
		return nil, fmt.Errorf("failed to merge skeleton %s and %s: %w", s.Ref, other.Ref, err)
	}

	return &Skeleton{
		Values:      values,
		Files:       MergeFiles(s.Files, other.Files),
		Description: other.Description,
		Ref:         other.Ref,
		// @FIXME(mohmann): this shouldn't be like this. due to the parameters
		// we should get rid of the merging logic and limit it to files only.
		Parameters: other.Parameters,
	}, nil
}
