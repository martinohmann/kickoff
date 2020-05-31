package git

import git "github.com/go-git/go-git/v5"

// Errors are aliased from go-git to avoid importing it just for error checks.
var (
	ErrRepositoryAlreadyExists = git.ErrRepositoryAlreadyExists
	ErrRepositoryNotExists     = git.ErrRepositoryNotExists
	NoErrAlreadyUpToDate       = git.NoErrAlreadyUpToDate
)
