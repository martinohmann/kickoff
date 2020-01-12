package skeleton

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/martinohmann/kickoff/pkg/file"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// Repository is the interface for a skeleton repository.
type Repository interface {
	// Skeleton obtains the info for the skeleton with name or an error if the
	// skeleton does not exist within the repository.
	Skeleton(name string) (*Info, error)

	// Skeletons returns infos for all skeletons available in the repository.
	// Returns any error that may occur while traversing the directory.
	Skeletons() ([]*Info, error)
}

type initializer interface {
	init() error
}

// OpenRepository opens a repository and returns it. If url points to a remote
// repository it will be looked up in the local cache and reused if possible.
// If the repository is not in the cache it will be cloned. Open will
// automatically checkout branches provided in the url. Returns any errors that
// occur while parsing the url opening the repository directory or during git
// actions.
func OpenRepository(url string) (Repository, error) {
	return openNamedRepository("", url)
}

func openNamedRepository(name, url string) (Repository, error) {
	info, err := ParseRepositoryURL(url)
	if err != nil {
		return nil, err
	}

	info.Name = name

	var r Repository

	switch {
	case info.Local && info.Branch == "":
		r = newLocalDir(info)
	case info.Local:
		r = newLocalRepo(info)
	default:
		r = newRemoteRepo(info)
	}

	if ri, ok := r.(initializer); ok {
		err = ri.init()
		if err != nil {
			return nil, err
		}
	}

	return r, nil
}

type localDir struct {
	info *RepositoryInfo
}

func newLocalDir(info *RepositoryInfo) *localDir {
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

func (r *localDir) Skeleton(name string) (*Info, error) {
	path := filepath.Join(r.info.LocalPath(), name)

	ok, err := isSkeletonDir(path)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, fmt.Errorf("skeleton %q not found in %s", name, r.info)
	}

	info := &Info{
		Name: name,
		Path: path,
		Repo: r.info,
	}

	return info, nil
}

func (r *localDir) Skeletons() ([]*Info, error) {
	return findSkeletons(r.info, r.info.LocalPath())
}

type localRepo struct {
	*localDir
}

func newLocalRepo(info *RepositoryInfo) *localRepo {
	return &localRepo{
		localDir: &localDir{info},
	}
}

func (r *localRepo) init() error {
	localPath := r.info.LocalPath()

	repo, err := git.PlainOpen(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local repository %s: %v", localPath, err)
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

func newRemoteRepo(info *RepositoryInfo) *remoteRepo {
	return &remoteRepo{
		localRepo: newLocalRepo(info),
	}
}

func (r *remoteRepo) init() error {
	localPath := r.info.LocalPath()

	log.WithField("url", r.info.String()).Debug("using remote skeleton repository")

	repo, err := git.PlainOpen(localPath)
	if err == git.ErrRepositoryNotExists {
		parentDir := filepath.Dir(localPath)

		err = os.MkdirAll(parentDir, 0755)
		if err != nil {
			return err
		}

		log.WithField("localPath", localPath).Debug("cloning remote skeleton repository")

		repo, err = git.PlainClone(localPath, false, &git.CloneOptions{
			URL: r.info.String(),
		})
		if err != nil {
			return fmt.Errorf("failed to clone repository %s: %v", r.info, err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to open local repository %s: %v", localPath, err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return err
	}

	status, err := wt.Status()
	if err != nil {
		return err
	}

	if !status.IsClean() {
		log.WithField("localPath", localPath).Debug("cleaning repository")

		err = wt.Clean(&git.CleanOptions{Dir: true})
		if err != nil {
			return fmt.Errorf("failed to clean repository at %s: %v", localPath, err)
		}
	}

	err = checkoutBranch(wt, r.info.Branch)
	if err != nil {
		return err
	}

	log.WithField("branch", r.info.Branch).Debug("pulling branch")

	err = wt.Pull(&git.PullOptions{
		SingleBranch: true,
		Depth:        1,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to pull branch %q: %v", r.info.Branch, err)
	}

	return nil
}

func checkoutBranch(wt *git.Worktree, branch string) error {
	log.WithField("branch", branch).Debug("checking out branch")

	err := wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
	})
	if err != nil {
		return fmt.Errorf("failed to checkout branch %q: %v", branch, err)
	}

	return nil
}
