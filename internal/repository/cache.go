package repository

import "sync"

var repoCache cache = &realCache{}

// EnableCache enables the repository cache. An enabled cache will cause New
// and NewNamed to return the same Repository instance for consecutive calls
// with pairs of the same name and url. This can speed up operations on remote
// skeleton repositories as it reduce the number of git operations that need to
// be carried out. The speedup may be noticeable when working with skeletons
// that have parents.
func EnableCache() {
	repoCache = &realCache{}
}

// DisableCache disables the repository cache.
func DisableCache() {
	repoCache = &nopCache{}
}

type cacheKey struct {
	Name string
	URL  string
}

type cache interface {
	get(key cacheKey) (Repository, bool)
	set(key cacheKey, repo Repository)
}

type realCache struct {
	sync.Mutex
	repos map[cacheKey]Repository
}

func (c *realCache) get(key cacheKey) (Repository, bool) {
	c.Lock()
	defer c.Unlock()
	repo, ok := c.repos[key]
	return repo, ok
}

func (c *realCache) set(key cacheKey, repo Repository) {
	c.Lock()
	if c.repos == nil {
		c.repos = make(map[cacheKey]Repository)
	}
	c.repos[key] = repo
	defer c.Unlock()
}

type nopCache struct{}

func (nopCache) get(key cacheKey) (Repository, bool) { return nil, false }
func (nopCache) set(key cacheKey, repo Repository)   {}
