package kickoff

import (
	"io"
	"os"
)

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
