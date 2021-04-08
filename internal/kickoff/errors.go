package kickoff

import (
	"errors"
	"fmt"
)

// ErrMergeEmpty is returned by MergeSkeletons if no skeletons were passed.
var ErrMergeEmpty = errors.New("cannot merge empty list of skeletons")

// Base validation errors.
var (
	invalidParentRef     = "invalid parent ref"
	invalidProjectConfig = "invalid project config"
	invalidRepositoryRef = "invalid repository ref"
	invalidSkeletonRef   = "invalid skeleton ref"
)

// ValidationError wraps all errors that occur during validation.
type ValidationError struct {
	Context string
	Err     error
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("%s: %s", e.Context, e.Err.Error())
	}
	return e.Err.Error()
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

func newValidationError(context, format string, args ...interface{}) *ValidationError {
	return &ValidationError{
		Context: context,
		Err:     fmt.Errorf(format, args...),
	}
}

func newProjectConfigError(format string, args ...interface{}) *ValidationError {
	return newValidationError(invalidProjectConfig, format, args...)
}

func newRepositoryRefError(format string, args ...interface{}) *ValidationError {
	return newValidationError(invalidRepositoryRef, format, args...)
}

func newParentRefError(format string, args ...interface{}) *ValidationError {
	return newValidationError(invalidParentRef, format, args...)
}

func newSkeletonRefError(format string, args ...interface{}) *ValidationError {
	return newValidationError(invalidSkeletonRef, format, args...)
}
