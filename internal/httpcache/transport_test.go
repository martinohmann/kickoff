package httpcache

import (
	"net/http"
	"testing"
	"time"

	"github.com/martinohmann/httpcache"
	"github.com/stretchr/testify/require"
)

type fakeTransport struct {
	capturedReq *http.Request
	resp        *http.Response
	err         error
}

func newFakeTransport() *fakeTransport {
	return &fakeTransport{
		resp: &http.Response{
			Header: http.Header{
				httpcache.XFromCache: {"1"},
			},
		},
	}
}

func (t *fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.capturedReq = req
	if t.err != nil {
		return nil, t.err
	}
	return t.resp, nil
}

func TestStaleIfErrorTransport(t *testing.T) {
	require := require.New(t)

	t.Run("adds stale-if-error to Cache-Control header", func(t *testing.T) {
		fakeTransport := newFakeTransport()
		trans := newStaleIfErrorTransport(fakeTransport, 0)

		req := &http.Request{Header: http.Header{}}

		expected := http.Header{"Cache-Control": {"stale-if-error"}}

		_, err := trans.RoundTrip(req)
		require.NoError(err)

		require.Equal(expected, fakeTransport.capturedReq.Header)
	})

	t.Run("adds stale-if-error=60 to Cache-Control header", func(t *testing.T) {
		fakeTransport := newFakeTransport()
		trans := newStaleIfErrorTransport(fakeTransport, 60*time.Second)

		req := &http.Request{Header: http.Header{}}

		expected := http.Header{"Cache-Control": {"stale-if-error=60"}}

		_, err := trans.RoundTrip(req)
		require.NoError(err)

		require.Equal(expected, fakeTransport.capturedReq.Header)
	})

	t.Run("honors existing 'stale-if-error' Cache-Control header value", func(t *testing.T) {
		fakeTransport := newFakeTransport()
		trans := newStaleIfErrorTransport(fakeTransport, 30*time.Second)

		req := &http.Request{Header: http.Header{
			"Cache-Control": {"stale-if-error=60"},
		}}

		_, err := trans.RoundTrip(req)
		require.NoError(err)

		require.Equal(req, fakeTransport.capturedReq)
	})

	t.Run("honors existing 'no-cache' Cache-Control header value", func(t *testing.T) {
		fakeTransport := newFakeTransport()
		trans := newStaleIfErrorTransport(fakeTransport, 30*time.Second)

		req := &http.Request{Header: http.Header{
			"Cache-Control": {"no-cache"},
		}}

		_, err := trans.RoundTrip(req)
		require.NoError(err)

		require.Equal(req, fakeTransport.capturedReq)
	})
}
