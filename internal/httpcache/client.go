package httpcache

import (
	"net/http"

	"github.com/kirsle/configdir"
	"github.com/martinohmann/httpcache"
	"github.com/martinohmann/httpcache/diskcache"
)

var cacheDir = configdir.LocalCache("kickoff", "httpcache")

// NewClient creates a new *http.Client which caches http responses on disk so
// that certain operations can be performed without an active internet
// connection if there's at least a stale cached response for an http request.
func NewClient() *http.Client {
	cache := diskcache.New(cacheDir)

	return NewClientWithCache(cache)
}

// NewClientWithCache creates a new *http.Client which uses cache for response
// caching.
func NewClientWithCache(cache httpcache.Cache) *http.Client {
	transport := httpcache.NewTransport(cache)

	return &http.Client{
		Transport: newStaleIfErrorTransport(transport, 0),
	}
}
