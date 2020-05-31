package git

import (
	"context"

	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/mock"
)

// FakeClient is a fake git client with can be used in tests.
type FakeClient struct {
	mock.Mock
}

// Clone implements Client.
func (c *FakeClient) Clone(ctx context.Context, url, localPath string) (Repository, error) {
	args := c.Called(ctx, url, localPath)
	if r, ok := args.Get(0).(Repository); ok {
		return r, args.Error(1)
	}
	return nil, args.Error(1)
}

// Open implements Client.
func (c *FakeClient) Open(path string) (Repository, error) {
	args := c.Called(path)
	if r, ok := args.Get(0).(Repository); ok {
		return r, args.Error(1)
	}
	return nil, args.Error(1)
}

// Init implements Client.
func (c *FakeClient) Init(path string) (Repository, error) {
	args := c.Called(path)
	if r, ok := args.Get(0).(Repository); ok {
		return r, args.Error(1)
	}
	return nil, args.Error(1)
}

// FakeRepository is a fake git repository which can be used in tests.
type FakeRepository struct {
	mock.Mock
}

// Fetch implements Repository.
func (r *FakeRepository) Fetch(ctx context.Context, refSpecs ...config.RefSpec) error {
	args := r.Called(ctx, refSpecs)
	return args.Error(0)
}

// ResolveRevision implements Repository.
func (r *FakeRepository) ResolveRevision(revision plumbing.Revision) (*plumbing.Hash, error) {
	args := r.Called(revision)
	if hash, ok := args.Get(0).(*plumbing.Hash); ok {
		return hash, args.Error(1)
	}
	return nil, args.Error(1)
}

// Checkout implements Repository.
func (r *FakeRepository) Checkout(hash plumbing.Hash) error {
	args := r.Called(hash)
	return args.Error(0)
}
