// Package git provides an abstraction over github.com/go-git/go-git/v5 to make
// it easy to mock out remote git operations in tests.
package git

import (
	"context"

	git "github.com/go-git/go-git/v5"
)

// Client is the interface for a client that can perform actions on git
// repositories.
type Client interface {
	// Clone a repository at url into localPath. The clone is performed
	// non-bare, so the repository will have a worktree. If the local path is
	// not empty ErrRepositoryAlreadyExists is returned.
	Clone(ctx context.Context, url, localPath string) (Repository, error)

	// Open opens a repository from the given path. It detects if the
	// repository is bare or a normal one. If the path doesn't contain a valid
	// repository ErrRepositoryNotExists is returned.
	Open(path string) (Repository, error)

	// Init creates an empty non-bare git repository at the given path.
	// Non-bare means that the repository will have worktree. If the path is
	// not empty ErrRepositoryAlreadyExists is returned.
	Init(path string) (Repository, error)
}

// NewClient creates a new Client which will perform real git operations on
// disk.
func NewClient() Client {
	return &client{}
}

type client struct{}

func (*client) Clone(ctx context.Context, url, localPath string) (Repository, error) {
	r, err := git.PlainCloneContext(ctx, localPath, false, &git.CloneOptions{
		URL: url,
	})
	if err != nil {
		return nil, err
	}

	return NewRepository(r), nil
}

func (*client) Open(path string) (Repository, error) {
	r, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}

	return NewRepository(r), nil
}

func (*client) Init(path string) (Repository, error) {
	r, err := git.PlainInit(path, false)
	if err != nil {
		return nil, err
	}

	return NewRepository(r), nil
}
