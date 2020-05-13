package skeleton

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/kirsle/configdir"
	"github.com/martinohmann/kickoff/internal/homedir"
)

// Info holds information about a skeleton.
type Info struct {
	Name string
	Path string
	Repo *RepositoryInfo
}

// String implements fmt.Stringer.
func (i *Info) String() string {
	if i.Repo == nil || len(i.Repo.Name) == 0 {
		return i.Name
	}

	return fmt.Sprintf("%s:%s", i.Repo.Name, i.Name)
}

// Config loads the skeleton config for the info.
func (i *Info) LoadConfig() (Config, error) {
	configPath := filepath.Join(i.Path, ConfigFileName)

	return LoadConfig(configPath)
}

var (
	// LocalCache holds the path to the local repository cache. This is platform
	// specific.
	LocalCache = configdir.LocalCache("kickoff", "repositories")
)

const (
	defaultRevision = "master"
)

// RepositoryInfo holds information about a skeleton repository.
type RepositoryInfo struct {
	Local    bool   `json:"local"`
	Name     string `json:"name"`
	Revision string `json:"revision,omitempty"`
	Path     string `json:"path"`
	Scheme   string `json:"scheme,omitempty"`
	User     string `json:"user,omitempty"`
	Host     string `json:"host,omitempty"`
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

	revision := i.Revision
	if revision == "" {
		revision = defaultRevision
	}

	revision = url.PathEscape(revision)

	return filepath.Join(LocalCache, i.Host, fmt.Sprintf("%s@%s", i.Path, revision))
}

// SkeletonsDir returns the local path of the skeletons dir within the
// repository.
func (i *RepositoryInfo) SkeletonsDir() string {
	return filepath.Join(i.LocalPath(), "skeletons")
}

// Validate validates a repository info. Returns an error if i does not
// describe a valid skeleton repository.
func (i *RepositoryInfo) Validate() error {
	fi, err := os.Stat(i.SkeletonsDir())
	if os.IsNotExist(err) || !fi.IsDir() {
		return fmt.Errorf("%s is not a valid skeleton repository, skeletons/ is not a directory", i.LocalPath())
	}

	return err
}

// ParseRepositoryURL parses a raw repository url into a repository info.
func ParseRepositoryURL(rawurl string) (*RepositoryInfo, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	if len(u.Path) == 0 {
		return nil, fmt.Errorf("unable to parse path from raw url %s", rawurl)
	}

	info := &RepositoryInfo{}

	if u.Host == "" {
		path, err := homedir.Expand(u.Path)
		if err != nil {
			return nil, err
		}

		abspath, err := filepath.Abs(path)
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

		revision, ok := u.Query()["revision"]
		if ok && len(revision) > 0 {
			info.Revision = revision[0]
		}

		if info.Revision == "" {
			info.Revision = defaultRevision
		}
	}

	return info, nil
}
