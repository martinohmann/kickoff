package httpcache

import (
	"fmt"
	"net/http"
	"time"

	"github.com/apex/log"
	"github.com/martinohmann/httpcache"
)

var logger = log.WithField("component", "httpcache")

type staleIfErrorTransport struct {
	http.RoundTripper
	headerValue string
}

// newStaleIfErrorTransport adds `stale-if-error=delta` to the `Cache-Control`
// header of every request the passes through it. It will not update requests
// that explicitly define the `stale-if-error` field or the have `no-cache` set
// in the `Cache-Control` header. HTTP requests are cloned before modification.
// It passes every HTTP request to the underlying http.RoundTripper.
func newStaleIfErrorTransport(rt http.RoundTripper, delta time.Duration) http.RoundTripper {
	var headerValue string

	if delta > 0 {
		headerValue = fmt.Sprint(int64(delta / time.Second))
	}

	return &staleIfErrorTransport{
		RoundTripper: rt,
		headerValue:  headerValue,
	}
}

func (t *staleIfErrorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	cch := parseCacheControlHeader(req.Header)

	if !cch.Has("stale-if-error") && !cch.Has("no-cache") {
		cch.Set("stale-if-error", t.headerValue)

		clonedReq := req.Clone(req.Context())
		clonedReq.Header.Set("Cache-Control", cch.String())

		req = clonedReq
	}

	logger.WithFields(log.Fields{
		"req.url":    req.URL,
		"req.method": req.Method,
	}).Debug("requesting resource")

	resp, err := t.RoundTripper.RoundTrip(req)
	if err == nil && resp.Header.Get(httpcache.XFromCache) == "1" {
		logger.WithFields(log.Fields{
			"req.url":    req.URL,
			"req.method": req.Method,
		}).Debug("got response from cache")
	}

	return resp, err
}
