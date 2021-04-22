package kickoff

// Repository is the interface for a skeleton repository.
type Repository interface {
	// GetSkeleton retrieves information about a skeleton from the repository.
	// Returns an error of type SkeletonNotFoundError if the named skeleton was
	// not found in the repository.
	GetSkeleton(name string) (*SkeletonRef, error)

	// ListSkeletons retrieves information about all skeletons in the
	// repository. If the repository is empty, ListSkeletons will return an
	// empty slice.
	ListSkeletons() ([]*SkeletonRef, error)

	// LoadSkeleton loads the skeleton with name from given repository. Returns
	// an error if loading the skeleton fails.
	LoadSkeleton(name string) (*Skeleton, error)

	// CreateSkeleton creates a new skeleton with name in the referenced
	// repository. Skeleton creation will fail with an error if ref does not
	// reference a local repository. The created skeleton contains an example
	// .kickoff.yaml and example README.md.skel as starter. Returns an error if
	// creating path or writing any of the files fails.
	CreateSkeleton(name string) (*SkeletonRef, error)
}

// Defaulter can set defaults for unset fields.
type Defaulter interface {
	// ApplyDefaults sets unset fields of the data structure to its default
	// values which might not necessarily be the zero value.
	ApplyDefaults()
}

// Validator can validate itself to ensure the absence of invalid values.
type Validator interface {
	// Validate returns an error if the data structure contains invalid values.
	Validate() error
}
