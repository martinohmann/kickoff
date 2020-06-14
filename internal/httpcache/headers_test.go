package httpcache

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCacheControlHeader(t *testing.T) {
	require := require.New(t)
	h := newCacheControlHeader()

	require.False(h.Has("public"))
	require.Empty(h.String())

	h.Set("public", "")

	require.True(h.Has("public"))
	require.Equal("public", h.String())

	h.Set("stale-if-error", "60")

	require.Equal("public, stale-if-error=60", h.String())
}

func TestParseCacheControlHeader(t *testing.T) {
	header := http.Header{"Cache-Control": {"public, stale-if-error=60"}}

	expected := cacheControlHeader{
		m: map[string]string{
			"public":         "",
			"stale-if-error": "60",
		},
		keys: []string{"public", "stale-if-error"},
	}

	result := parseCacheControlHeader(header)

	require.Equal(t, expected, result)

	header = http.Header{"Cache-Control": {""}}

	result = parseCacheControlHeader(header)

	require.Equal(t, newCacheControlHeader(), result)
}
