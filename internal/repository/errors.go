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

// Error implements the error interface.
func (e SkeletonNotFoundError) Error() string {
	if e.RepoName == "" {
		return fmt.Sprintf("skeleton %q not found", e.Name)
	}

	return fmt.Sprintf("skeleton %q not found in repository %q", e.Name, e.RepoName)
}

type RevisionNotFoundError struct {
	RepoRef kickoff.RepoRef
}

// Error implements the error interface.
func (e RevisionNotFoundError) Error() string {
	repo := e.RepoRef.Name
	if repo == "" {
		repo = e.RepoRef.URL
	}

	return fmt.Sprintf("revision %q not found in repository %q", e.RepoRef.Revision, repo)
}

type InvalidSkeletonRepositoryError struct {
	RepoRef kickoff.RepoRef
}

// Error implements the error interface.
func (e InvalidSkeletonRepositoryError) Error() string {
	repo := e.RepoRef.Name
	if repo == "" {
		repo = e.RepoRef.String()
	}

	return fmt.Sprintf("%q is not a valid skeleton repository", repo)
}
