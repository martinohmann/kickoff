package homedir

import (
	"os"
	"testing"

	"github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/assert"
)

func TestCollapse(t *testing.T) {
	testCases := []homedirTestCase{
		{
			name:     "full path gets collapsed",
			home:     "/home/user",
			path:     "/home/user/foo",
			expected: "~/foo",
		},
		{
			name:     "full path gets collapsed #2",
			home:     "/home/user",
			path:     "/home/user/foo/",
			expected: "~/foo/",
		},
		{
			name:     "full path gets collapsed (homedir with trailing slash)",
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
			name:     "already collapsed path is returned as is",
			home:     "/home/user",
			path:     "~/foo/bar",
			expected: "~/foo/bar",
		},
		{
			name:     "homedir of another user with the same prefix does not get collapsed",
			home:     "/home/user",
			path:     "/home/userfoo/dir",
			expected: "/home/userfoo/dir",
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

	runTestCases(t, testCases, Collapse)
}

func TestExpand(t *testing.T) {
	testCases := []homedirTestCase{
		{
			name:     "full path gets expanded",
			home:     "/home/user",
			path:     "~/foo",
			expected: "/home/user/foo",
		},
		{
			name:     "full path with trailing slash gets expanded",
			home:     "/home/user/",
			path:     "~/foo",
			expected: "/home/user/foo",
		},
		{
			name:     "home gets expanded",
			home:     "/home/user",
			path:     "~",
			expected: "/home/user",
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

	runTestCases(t, testCases, Expand)
}

type homedirTestCase struct {
	name     string
	home     string
	path     string
	expected string
}

func runTestCases(t *testing.T, testCases []homedirTestCase, fn func(path string) string) {
	defer disableCache()()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer overrideHome(tc.home)()

			assert.Equal(t, tc.expected, fn(tc.path))
		})
	}
}

func overrideHome(path string) func() {
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", path)

	return func() { os.Setenv("HOME", originalHome) }
}

func disableCache() func() {
	originalDisableCache := homedir.DisableCache
	homedir.DisableCache = true

	return func() { homedir.DisableCache = originalDisableCache }
}
