package gitignore

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	var tests = []struct {
		name        string
		handler     func(t *testing.T) http.HandlerFunc
		expected    []string
		expectError bool
		expectedErr error
	}{
		{
			name:     "returns a list of gitignores",
			expected: []string{"foo", "bar", "baz", "qux"},
			handler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					if r.RequestURI != "/list" {
						t.Fatalf("unexpected request URI: %s", r.RequestURI)
					}

					_, err := w.Write([]byte("foo,bar\nbaz\nqux\n"))
					if err != nil {
						t.Fatal(err)
					}

					w.WriteHeader(200)
				}
			},
		},
		{
			name:        "returns error on non-200 status codes",
			expectError: true,
			expectedErr: errors.New("gitignore.io returned status code 500 while listing gitignore templates"),
			handler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(500)
				}
			},
		},
		{
			name:        "returns connection errors",
			expectError: true,
			handler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					panic("whoops")
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			done := testServer(test.handler(t))
			defer done()

			gitignores, err := List(context.Background())
			if test.expectError {
				require.Error(t, err)

				if test.expectedErr != nil {
					assert.Equal(t, test.expectedErr, err)
				}
			} else {
				require.NoError(t, err)

				assert.Equal(t, test.expected, gitignores)
			}
		})
	}
}

func TestGet(t *testing.T) {
	var tests = []struct {
		name        string
		query       string
		handler     func(t *testing.T) http.HandlerFunc
		expected    string
		expectError bool
		expectedErr error
	}{
		{
			name:     "returns a gitignore, with space trimmed",
			query:    "go,python",
			expected: "coverage.txt\nvendor/\n__pycache__",
			handler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					if r.RequestURI != "/go,python" {
						t.Fatalf("unexpected request URI: %s", r.RequestURI)
					}

					_, err := w.Write([]byte("\ncoverage.txt\nvendor/\n__pycache__\n"))
					if err != nil {
						t.Fatal(err)
					}

					w.WriteHeader(200)
				}
			},
		},
		{
			name:        "returns error on non-200 status codes",
			expectError: true,
			expectedErr: errors.New("gitignore.io returned status code 500 while fetching gitignore template"),
			handler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(500)
				}
			},
		},
		{
			name:        "returns well-known 404 error",
			expectError: true,
			expectedErr: ErrNotFound,
			handler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(404)
				}
			},
		},
		{
			name:        "returns connection errors",
			expectError: true,
			handler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					panic("whoops")
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			done := testServer(test.handler(t))
			defer done()

			gitignores, err := Get(context.Background(), test.query)
			if test.expectError {
				require.Error(t, err)

				if test.expectedErr != nil {
					assert.Equal(t, test.expectedErr, err)
				}
			} else {
				require.NoError(t, err)

				assert.Equal(t, test.expected, gitignores)
			}
		})
	}
}

func testServer(handler http.HandlerFunc) func() {
	server := httptest.NewServer(handler)
	originalAPIBaseURL := apiBaseURL
	apiBaseURL = server.URL

	return func() {
		apiBaseURL = originalAPIBaseURL
		server.Close()
	}
}
