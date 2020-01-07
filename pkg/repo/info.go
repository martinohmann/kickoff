package repo

import (
	"net/url"
	"path/filepath"
	"strings"

	"github.com/kirsle/configdir"
)

var (
	// LocalCache holds the path to the local repository cache. This is platform
	// specific.
	LocalCache = configdir.LocalCache("kickoff", "repositories")
)

// Info holds information about a skeleton repository.
type Info struct {
	Local  bool
	Branch string
	Path   string
	Scheme string
	User   string
	Host   string
}

// String implements fmt.Stringer.
func (i *Info) String() string {
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
func (i *Info) LocalPath() string {
	if i.Local {
		return i.Path
	}

	return filepath.Join(LocalCache, i.Host, i.Path)
}

// ParseURL parses a raw repository url into a repository info.
func ParseURL(rawurl string) (*Info, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	info := &Info{}

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
