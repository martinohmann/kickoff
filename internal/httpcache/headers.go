package httpcache

import (
	"net/http"
	"strings"
)

type cacheControlHeader struct {
	m    map[string]string
	keys []string
}

func newCacheControlHeader() cacheControlHeader {
	return cacheControlHeader{m: make(map[string]string)}
}

func (c *cacheControlHeader) Has(key string) bool {
	_, ok := c.Get(key)
	return ok
}

func (c *cacheControlHeader) Get(key string) (string, bool) {
	v, ok := c.m[key]
	return v, ok
}

func (c *cacheControlHeader) Set(key, value string) {
	_, exists := c.Get(key)
	if !exists {
		c.keys = append(c.keys, key)
	}
	c.m[key] = value
}

func (c *cacheControlHeader) String() string {
	var sb strings.Builder

	for i, k := range c.keys {
		v := c.m[k]
		sb.WriteString(k)
		if v != "" {
			sb.WriteRune('=')
			sb.WriteString(v)
		}

		if i < len(c.keys)-1 {
			sb.WriteString(", ")
		}
	}

	return sb.String()
}

func parseCacheControlHeader(headers http.Header) cacheControlHeader {
	cc := newCacheControlHeader()

	header := headers.Get("Cache-Control")
	for _, key := range strings.Split(header, ",") {
		key = strings.Trim(key, " ")
		if key == "" {
			continue
		}

		var val string

		if strings.ContainsRune(key, '=') {
			key, val = splitKeyValue(key)
		}

		cc.Set(key, val)
	}

	return cc
}

func splitKeyValue(s string) (string, string) {
	kv := strings.Split(s, "=")
	return strings.Trim(kv[0], " "), strings.Trim(kv[1], ",")
}
