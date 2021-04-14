package cmdutil

import "fmt"

type RepositoryAlreadyExistsError string

// Error implements the error interface.
func (e RepositoryAlreadyExistsError) Error() string {
	return fmt.Sprintf("repository %q already exists", string(e))
}

type RepositoryNotConfiguredError string

// Error implements the error interface.
func (e RepositoryNotConfiguredError) Error() string {
	return fmt.Sprintf("repository %q not configured", string(e))
}
