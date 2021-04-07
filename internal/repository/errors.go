package repository

import (
	"errors"
	"fmt"

	"github.com/martinohmann/kickoff/internal/kickoff"
)

var (
	// ErrNoRepositories is returned by NewMultiRepository if no repositories are
	// configured.
	ErrNoRepositories = errors.New("no skeleton repositories configured")

	// ErrNotARemoteRepository is returned by NewRemoteRepository if the provided
	// info does not describe a remote repository.
	ErrNotARemoteRepository = errors.New("not a remote repository")
)

// SkeletonNotFoundError is the error returned if a skeleton cannot be found in
// a repository.
type SkeletonNotFoundError struct {
	Name     string
	RepoName string
}

// Error implements error.
func (e SkeletonNotFoundError) Error() string {
	if e.RepoName == "" {
		return fmt.Sprintf("skeleton %q not found", e.Name)
	}

	return fmt.Sprintf("skeleton %q not found in repository %q", e.Name, e.RepoName)
}

// DependencyCycleError is the error returned while loading a skeleton's parent
// if a dependency cycle is detected.
type DependencyCycleError struct {
	ParentRef kickoff.ParentRef
}

// Error implements error.
func (e DependencyCycleError) Error() string {
	return fmt.Sprintf("dependency cycle detected for parent: %#v", e.ParentRef)
}
