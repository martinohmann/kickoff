package skeleton

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/kirsle/configdir"
	"github.com/martinohmann/kickoff/pkg/config"
)

// Info holds information about a skeleton.
type Info struct {
	Name string
	Path string
}

// Config loads the skeleton config for the info.
func (i *Info) Config() (config.Skeleton, error) {
	return config.LoadSkeleton(filepath.Join(i.Path, config.SkeletonConfigFile))
}

// Walk recursively walks all files and directories of the skeleton. This
// behaves exactly as filepath.Walk, except that it will ignore the skeleton's
// ConfigFile.
func (i *Info) Walk(walkFn filepath.WalkFunc) error {
	return filepath.Walk(i.Path, func(path string, info os.FileInfo, err error) error {
		if info.Name() == config.SkeletonConfigFile {
			// ignore skeleton config file
			return err
		}

		return walkFn(path, info, err)
	})
}

var (
	// LocalCache holds the path to the local repository cache. This is platform
	// specific.
	LocalCache = configdir.LocalCache("kickoff", "repositories")
)

// RepositoryInfo holds information about a skeleton repository.
type RepositoryInfo struct {
	Local  bool
	Branch string
	Path   string
	Scheme string
	User   string
	Host   string
}

// String implements fmt.Stringer.
func (i *RepositoryInfo) String() string {
	if i.Local {
		return i.Path
	}

	var sb strings.Builder

	sb.WriteString(i.Scheme)
	sb.WriteString("://")

	if i.User != "" {
		sb.WriteString(i.User)
		sb.WriteByte('@')
	}

	sb.WriteString(i.Host)
	sb.WriteByte('/')
	sb.WriteString(i.Path)

	return sb.String()
}

// LocalPath returns the local path to a repository. For local repositories
// this is just the actual path on disk, for remote repositories this returns a
// path within the LocalCache directory.
func (i *RepositoryInfo) LocalPath() string {
	if i.Local {
		return i.Path
	}

	return filepath.Join(LocalCache, i.Host, i.Path)
}

// ParseRepositoryURL parses a raw repository url into a repository info.
func ParseRepositoryURL(rawurl string) (*RepositoryInfo, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	info := &RepositoryInfo{}

	branch, ok := u.Query()["branch"]
	if ok && len(branch) > 0 {
		info.Branch = branch[0]
	}

	if u.Host == "" {
		abspath, err := filepath.Abs(u.Path)
		if err != nil {
			return nil, err
		}

		info.Local = true
		info.Path = abspath
	} else {
		info.Scheme = u.Scheme
		info.User = u.User.String()
		info.Host = u.Host
		info.Path = u.Path[1:]
	}

	return info, nil
}
