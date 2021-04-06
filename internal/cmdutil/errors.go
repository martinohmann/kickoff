package cmdutil

import (
	"errors"
	"fmt"
)

var (
	// ErrEmptyOutputDir is returned if the user passed an empty string as the
	// output directory.
	ErrEmptyOutputDir = errors.New("output dir must not be an empty string")

	// ErrInvalidOutputFormat is returned if the output format flag contains an
	// invalid value.
	ErrInvalidOutputFormat = errors.New("--output must be 'yaml' or 'json'")
)

type RepositoryNotConfiguredError struct {
	Name string
}

// Error implements the error interface.
func (e RepositoryNotConfiguredError) Error() string {
	return fmt.Sprintf("repository %q not configured", e.Name)
}
