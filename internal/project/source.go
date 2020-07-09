package project

import (
	"io"
	"os"
	"path/filepath"

	"github.com/martinohmann/kickoff/internal/file"
)

// Source is the interface for a source file that should be created in new
// projects.
type Source interface {
	// Path must return the path relative to the new project root.
	Path() string

	// Mode returns the mode of the file. The target in the project directory
	// will be created using this mode. The mode is also used to determine if
	// the file is regular or a directory.
	Mode() os.FileMode

	// Reader provides an io.Reader to read the contents of the file.
	Reader() (io.Reader, error)

	// IsTemplate returns to if the source is a template.
	IsTemplate() bool
}

type source struct {
	reader io.Reader
	mode   os.FileMode
	path   string
}

// NewSource creates a new source from given reader, path and mode. The path
// must be relative to the new project root.
func NewSource(r io.Reader, path string, mode os.FileMode) Source {
	return &source{
		reader: r,
		path:   path,
		mode:   mode,
	}
}

func (s source) Path() string { return s.path }

func (s source) Mode() os.FileMode { return s.mode }

func (s source) Reader() (io.Reader, error) { return s.reader, nil }

func (s source) IsTemplate() bool {
	return !s.Mode().IsDir() && filepath.Ext(s.path) == ".skel"
}

// Destination describes the destination a project file should be written to.
type Destination struct {
	// Base is the base dir of the project.
	Base string
	// Path is the path relative to the base dir.
	Path string
}

// RelPath returns the path relative to the project root.
func (d Destination) RelPath() string {
	return d.Path
}

// AbsPath returns the absolute file path of the destination.
func (d Destination) AbsPath() string {
	return filepath.Join(d.Base, d.Path)
}

// Exists returns true if the destination already exists.
func (d Destination) Exists() bool {
	return file.Exists(d.AbsPath())
}
