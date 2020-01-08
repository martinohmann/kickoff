package repo

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/martinohmann/kickoff/pkg/file"
	"github.com/martinohmann/kickoff/pkg/skeleton"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// Repo is the interface for a skeleton repository.
type Repo interface {
	// Skeleton obtains the info for the skeleton with name or an error if the
	// skeleton does not exist within the repository.
	Skeleton(name string) (*skeleton.Info, error)

	// Skeletons returns infos for all skeletons available in the repository.
	// Returns any error that may occur while traversing the directory.
	Skeletons() ([]*skeleton.Info, error)

	// init is called after opening the repository.
	init() error
}

// Open opens a repository and returns it. If url points to a remote repository
// it will be looked up in the local cache and reused if possible. If the
// repository is not in the cache it will be cloned. Open will automatically
// checkout branches provided in the url. Returns any errors that occur while
// parsing the url opening the repository directory or during git actions.
func Open(url string) (Repo, error) {
	info, err := ParseURL(url)
	if err != nil {
		return nil, err
	}

	var r Repo

	switch {
	case info.Local && info.Branch == "":
		r = newLocalDir(info)
	case info.Local:
		r = newLocalRepo(info)
	default:
		r = newRemoteRepo(info)
	}

	err = r.init()
	if err != nil {
		return nil, err
	}

	return r, nil
}

type localDir struct {
	info *Info
}

func newLocalDir(info *Info) *localDir {
	return &localDir{
		info: info,
	}
}

func (r *localDir) init() error {
	path := r.info.LocalPath()

	ok, err := file.IsDirectory(path)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("%s is not a directory", path)
	}

	return nil
}

func (r *localDir) Skeleton(name string) (*skeleton.Info, error) {
	path := filepath.Join(r.info.LocalPath(), name)

	ok, err := isSkeletonDir(path)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, fmt.Errorf("skeleton %q not found in %s", name, r.info)
	}

	info := &skeleton.Info{
		Name: name,
		Path: path,
	}

	return info, nil
}

func (r *localDir) Skeletons() ([]*skeleton.Info, error) {
	return findSkeletons(r.info.LocalPath())
}

type localRepo struct {
	*localDir
}

func newLocalRepo(info *Info) *localRepo {
	return &localRepo{
		localDir: &localDir{info},
	}
}

func (r *localRepo) init() error {
	repo, err := git.PlainOpen(r.info.LocalPath())
	if err != nil {
		return err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return err
	}

	return checkoutBranch(wt, r.info.Branch)
}

type remoteRepo struct {
	*localRepo
}

func newRemoteRepo(info *Info) *remoteRepo {
	return &remoteRepo{
		localRepo: newLocalRepo(info),
	}
}

func (r *remoteRepo) init() error {
	localPath := r.info.LocalPath()

	repo, err := git.PlainOpen(localPath)
	if err == git.ErrRepositoryNotExists {
		parentDir := filepath.Dir(localPath)

		err = os.MkdirAll(parentDir, 0755)
		if err != nil {
			return err
		}

		var clonedRepo *git.Repository
		clonedRepo, err = git.PlainClone(localPath, false, &git.CloneOptions{
			URL: r.info.String(),
		})
		repo = clonedRepo
	}

	if err != nil {
		return err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return err
	}

	err = wt.Clean(&git.CleanOptions{})
	if err != nil {
		return err
	}

	err = checkoutBranch(wt, r.info.Branch)
	if err != nil {
		return err
	}

	return wt.Pull(&git.PullOptions{})
}

func checkoutBranch(wt *git.Worktree, branch string) error {
	return wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
	})
}
