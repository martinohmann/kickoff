package kickoff

import (
	"context"
	"io"
	"os"
)

// Repository is the interface for a skeleton repository.
type Repository interface {
	// GetSkeleton retrieves information about a skeleton from the
	// repository. The passed in context is propagated to all operations that
	// cross API boundaries (e.g. git operations) and can be used to enforce
	// timeouts or cancel them. Returns an error of type SkeletonNotFoundError
	// if the named skeleton was not found in the repository.
	GetSkeleton(ctx context.Context, name string) (*SkeletonRef, error)

	// ListSkeletons retrieves information about all skeletons in the
	// repository. The passed in context is propagated to all operations that
	// cross API boundaries (e.g. git operations) and can be used to enforce
	// timeouts or cancel them. If the repository is empty, ListSkeletons
	// will return an empty slice.
	ListSkeletons(ctx context.Context) ([]*SkeletonRef, error)
}

// Defaulter can set defaults for unset fields.
type Defaulter interface {
	// ApplyDefaults sets unset fields of the data structure to its default
	// values which might not necessarily be the zero value.
	ApplyDefaults()
}

// File is the interface for a file that should be created in new projects.
type File interface {
	// Path must return the path relative to the new project root.
	Path() string

	// Mode returns the mode of the file. The target in the project directory
	// will be created using this mode. The mode is also used to determine if
	// the file is regular or a directory.
	Mode() os.FileMode

	// Reader provides an io.Reader to read the contents of the file.
	Reader() (io.Reader, error)

	// IsTemplate returns true if the source is a template.
	IsTemplate() bool
}
