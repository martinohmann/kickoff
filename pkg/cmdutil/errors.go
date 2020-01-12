package cmdutil

import "errors"

var (
	// ErrEmptyOutputDir is returned if the user passed an empty string as the
	// output directory.
	ErrEmptyOutputDir = errors.New("output dir must not be an empty string")

	// ErrEmptySkeletonName is returned if the user passed an empty string as
	// skeleton name.
	ErrEmptySkeletonName = errors.New("skeleton name must be provided")

	// ErrInvalidOutputFormat is returned if the output format flag contains an
	// invalid value.
	ErrInvalidOutputFormat = errors.New("--output must be 'yaml' or 'json'")
)
