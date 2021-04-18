package repository

import (
	"errors"
	"fmt"

	"github.com/martinohmann/kickoff/internal/kickoff"
)

// ErrNoRepositories is returned by NewMultiRepository if no repositories are
// configured.
var ErrNoRepositories = errors.New("no skeleton repositories configured")

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

// SkeletonAlreadyExistsError is the error returned if a skeleton already
// exists in a repository.
type SkeletonAlreadyExistsError struct {
	Name     string
	RepoName string
}

// Error implements the error interface.
func (e SkeletonAlreadyExistsError) Error() string {
	if e.RepoName == "" {
		return fmt.Sprintf("skeleton %q already exists", e.Name)
	}

	return fmt.Sprintf("skeleton %q already exists in repository %q", e.Name, e.RepoName)
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
