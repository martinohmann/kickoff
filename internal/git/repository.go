package git

import (
	"context"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

// Repository is the interface for a git repository.
type Repository interface {
	// Fetch fetches references along with the objects necessary to complete
	// their histories, from the remote named origin.
	//
	// Returns nil if the operation is successful, NoErrAlreadyUpToDate if
	// there are no changes to be fetched, or an error.
	Fetch(ctx context.Context, refSpecs ...config.RefSpec) error

	// ResolveRevision resolves revision to corresponding hash. It will always
	// resolve to a commit hash, not a tree or annotated tag.
	ResolveRevision(revision plumbing.Revision) (*plumbing.Hash, error)

	// Checkout checks out the commit referenced by the provided hash. The
	// checkout is performed in force mode to throw away local changes.
	Checkout(hash plumbing.Hash) error
}

// NewRepository creates a new Repository from given go-git repository.
func NewRepository(repo *git.Repository) Repository {
	return &repository{repo}
}

type repository struct {
	*git.Repository
}

func (r *repository) Fetch(ctx context.Context, refSpecs ...config.RefSpec) error {
	return r.Repository.FetchContext(ctx, &git.FetchOptions{
		RefSpecs: refSpecs,
	})
}

func (r *repository) Checkout(hash plumbing.Hash) error {
	worktree, err := r.Worktree()
	if err != nil {
		return err
	}

	return worktree.Checkout(&git.CheckoutOptions{
		Hash:  hash,
		Force: true,
	})
}
