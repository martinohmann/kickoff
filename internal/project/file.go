package project

import (
	"bytes"
	"io"
	"os"
)

// File describes the minimal interface for a file that should be processed by the
// project builder.
type File interface {
	// Path must return the path relative to the new project root.
	Path() string

	// Mode returns the mode of the file. The target in the project directory
	// will be created using this mode. The mode is also used to determine if
	// the file is regular or a directory.
	Mode() os.FileMode

	// Reader provides an io.Reader to read the contents of the file.
	Reader() (io.Reader, error)
}

// fileInfo acts as a File but provides the file content from a buffer.
type fileInfo struct {
	relPath string
	content []byte
	mode    os.FileMode
}

// Path implements File.
func (f *fileInfo) Path() string {
	return f.relPath
}

// Path implements File.
func (f *fileInfo) Mode() os.FileMode {
	return f.mode
}

// Path implements File.
func (f *fileInfo) Reader() (io.Reader, error) {
	return bytes.NewReader(f.content), nil
}

// fileType returns the string representation of the file's type.
func fileType(f File) string {
	if f.Mode().IsDir() {
		return "directory"
	}

	return "file"
}

// Stats hold stats about the project creation result.
type Stats struct {
	DirsCreated  int
	FilesCreated int
	Skipped      int
}

func (s *Stats) increment(f File) {
	if f.Mode().IsDir() {
		s.DirsCreated++
	} else {
		s.FilesCreated++
	}
}
