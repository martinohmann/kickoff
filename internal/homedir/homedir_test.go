package homedir

import (
	"os"
	"testing"

	gohomedir "github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollapse(t *testing.T) {
	tests := []struct {
		name        string
		home        string
		path        string
		expected    string
		expectedErr error
	}{
		{
			name:     "full path gets collapsed",
			home:     "/home/user",
			path:     "/home/user/foo",
			expected: "~/foo",
		},
		{
			name:     "full path with trailing slash gets collapsed",
			home:     "/home/user/",
			path:     "/home/user/foo",
			expected: "~/foo",
		},
		{
			name:     "home gets collapsed",
			home:     "/home/user",
			path:     "/home/user",
			expected: "~",
		},
		{
			name:     "relative paths are left untouched",
			home:     "/home/user",
			path:     "./foo/bar",
			expected: "./foo/bar",
		},
		{
			name:     "absolute paths outside of home are left untouched",
			home:     "/home/user",
			path:     "/usr/local/bin/kickoff",
			expected: "/usr/local/bin/kickoff",
		},
	}

	originalDisableCache := gohomedir.DisableCache
	gohomedir.DisableCache = true
	defer func() {
		gohomedir.DisableCache = originalDisableCache
	}()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			restore := overrideHome(test.home)
			defer restore()

			collapsed, err := Collapse(test.path)
			if test.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, test.expectedErr, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected, collapsed)
			}
		})
	}
}

func overrideHome(path string) func() {
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", path)

	return func() {
		os.Setenv("HOME", originalHome)
	}
}
