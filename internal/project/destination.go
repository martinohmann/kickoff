package project

import (
	"os"
	"path/filepath"
)

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
	_, err := os.Stat(d.AbsPath())
	return err == nil
}
