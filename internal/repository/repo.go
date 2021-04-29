package repository

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/martinohmann/kickoff/internal/kickoff"
)

// Options for opening a repository.
type Options struct {
	// Fetcher is used to fetch remote repositories. If nil a default git
	// fetcher will be used.
	Fetcher RemoteFetcher
}

// Open opens a repository at url. Returns an error if url is not a valid local
// path or remote url.
func Open(ctx context.Context, url string, opts *Options) (kickoff.Repository, error) {
	return openNamed(ctx, "", url, opts)
}

// openNamed opens a named repository. The name is propagated into the
// repository ref that is attached to every skeleton that is retrieved from it.
// Apart from that is behaves exactly like Open.
func openNamed(ctx context.Context, name, url string, opts *Options) (kickoff.Repository, error) {
	ref, err := kickoff.ParseRepoRef(url)
	if err != nil {
		return nil, err
	}

	ref.Name = name

	return OpenRef(ctx, *ref, opts)
}

// OpenRef opens a repository from a repository reference. Ref may reference a
// local or remote repository.
func OpenRef(ctx context.Context, ref kickoff.RepoRef, opts *Options) (kickoff.Repository, error) {
	if err := ref.Validate(); err != nil {
		return nil, err
	}

	if opts == nil {
		opts = &Options{}
	}

	if opts.Fetcher == nil {
		opts.Fetcher = defaultFetcher
	}

	if ref.IsRemote() {
		if err := opts.Fetcher.FetchRemote(ctx, ref); err != nil {
			return nil, err
		}
	}

	return newRepository(ref)
}

// repository is a local skeleton repository. A local skeleton repository
// can be any directory on disk that contains a skeletons/ subdirectory.
type repository struct {
	ref kickoff.RepoRef
}

// newRepository creates a kickoff.Repository from ref. Returns an error if
// resolving the absolute path to the skeleton repository fails.
func newRepository(ref kickoff.RepoRef) (kickoff.Repository, error) {
	dir := ref.SkeletonsPath()

	fi, err := os.Stat(dir)
	if err != nil || !fi.IsDir() {
		return nil, InvalidSkeletonRepositoryError{RepoRef: ref}
	}

	return &repository{ref: ref}, nil
}

func (r *repository) GetSkeleton(name string) (*kickoff.SkeletonRef, error) {
	path := r.ref.SkeletonPath(name)

	if !isSkeletonDir(path) {
		return nil, SkeletonNotFoundError{name, r.ref.Name}
	}

	info := &kickoff.SkeletonRef{
		Name: name,
		Path: path,
		Repo: &r.ref,
	}

	return info, nil
}

func (r *repository) ListSkeletons() ([]*kickoff.SkeletonRef, error) {
	refs, err := listSkeletons(&r.ref, r.ref.SkeletonsPath())
	if err != nil {
		return nil, fmt.Errorf("failed to list skeletons: %w", err)
	}

	return refs, nil
}

func (r *repository) LoadSkeleton(name string) (*kickoff.Skeleton, error) {
	return loadSkeleton(r, name)
}

func (r *repository) CreateSkeleton(name string) (*kickoff.SkeletonRef, error) {
	err := createSkeleton(r.ref, name)
	if err != nil {
		return nil, err
	}

	return r.GetSkeleton(name)
}

func listSkeletons(repoRef *kickoff.RepoRef, dir string) ([]*kickoff.SkeletonRef, error) {
	refs := make([]*kickoff.SkeletonRef, 0)

	fileInfos, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, fi := range fileInfos {
		if !fi.IsDir() {
			continue
		}

		path := filepath.Join(dir, fi.Name())

		if isSkeletonDir(path) {
			abspath, err := filepath.Abs(path)
			if err != nil {
				return nil, err
			}

			refs = append(refs, &kickoff.SkeletonRef{
				Name: fi.Name(),
				Path: abspath,
				Repo: repoRef,
			})
			continue
		}

		skels, err := listSkeletons(repoRef, path)
		if err != nil {
			return nil, err
		}

		for _, s := range skels {
			refs = append(refs, &kickoff.SkeletonRef{
				Name: filepath.Join(fi.Name(), s.Name),
				Path: s.Path,
				Repo: repoRef,
			})
		}
	}

	return refs, nil
}

// isSkeletonDir returns true if dir is a skeleton dir. Skeleton dirs are
// detected by the fact that they contain a .kickoff.yaml file.
func isSkeletonDir(dir string) bool {
	configPath := filepath.Join(dir, kickoff.SkeletonConfigFileName)

	fi, err := os.Stat(configPath)
	if err != nil {
		return false
	}

	return fi.Mode().IsRegular()
}

// LoadSkeletons loads multiple skeletons from given repository. Returns an
// error if loading of any of the skeletons fails.
func LoadSkeletons(repo kickoff.Repository, names []string) ([]*kickoff.Skeleton, error) {
	skeletons := make([]*kickoff.Skeleton, len(names))

	for i, name := range names {
		skeleton, err := repo.LoadSkeleton(name)
		if err != nil {
			return nil, err
		}

		skeletons[i] = skeleton
	}

	return skeletons, nil
}

func loadSkeleton(repo kickoff.Repository, name string) (*kickoff.Skeleton, error) {
	ref, err := repo.GetSkeleton(name)
	if err != nil {
		return nil, err
	}

	config, err := ref.LoadConfig()
	if err != nil {
		return nil, err
	}

	files, err := loadSkeletonFiles(ref)
	if err != nil {
		return nil, err
	}

	s := &kickoff.Skeleton{
		Description: config.Description,
		Values:      config.Values,
		Ref:         ref,
		Files:       files,
	}

	return s, nil
}

func loadSkeletonFiles(ref *kickoff.SkeletonRef) ([]*kickoff.BufferedFile, error) {
	files := make([]*kickoff.BufferedFile, 0)

	err := filepath.Walk(ref.Path, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi.Name() == kickoff.SkeletonConfigFileName {
			// ignore dirs and the skeleton config file
			return nil
		}

		relPath, err := filepath.Rel(ref.Path, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			// ignore skeleton dir itself
			return nil
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		if fi.Mode().IsDir() {
			files = append(files, &kickoff.BufferedFile{
				RelPath:     relPath,
				Mode:        fi.Mode(),
				SkeletonRef: ref,
			})
			return nil
		}

		if !fi.Mode().IsRegular() {
			return fmt.Errorf("%s is not a regular file", absPath)
		}

		if fi.Size() > 100*1024*1024 {
			return fmt.Errorf("file %s too large: refusing to load files larger than 100 MiB", absPath)
		}

		buf, err := os.ReadFile(absPath)
		if err != nil {
			return err
		}

		files = append(files, &kickoff.BufferedFile{
			RelPath:     relPath,
			Content:     buf,
			Mode:        fi.Mode(),
			SkeletonRef: ref,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}
