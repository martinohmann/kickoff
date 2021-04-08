package kickoff

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseRepoRef(t *testing.T) {
	// override local user cache dir to be able to make test assertions on
	// paths.
	oldCacheDir := LocalRepositoryCacheDir
	LocalRepositoryCacheDir = "/home/someuser/.cache/kickoff/repositories"
	defer func() { LocalRepositoryCacheDir = oldCacheDir }()

	testCases := []struct {
		name     string
		s        string
		expected *RepoRef
		err      error
	}{
		{
			name: "local path",
			s:    "foo.bar.baz/johndoe/repo",
			expected: &RepoRef{
				Path: "foo.bar.baz/johndoe/repo",
			},
		},
		{
			name: "local abspath",
			s:    "/some/repo",
			expected: &RepoRef{
				Path: "/some/repo",
			},
		},
		{
			name: "local relpath",
			s:    "../../some/repo",
			expected: &RepoRef{
				Path: "../../some/repo",
			},
		},
		{
			name: "file protocol",
			s:    "file:///some/repo",
			expected: &RepoRef{
				Path: "/some/repo",
			},
		},
		{
			name: "url",
			s:    "https://foo.bar.baz/johndoe/repo",
			expected: &RepoRef{
				URL:      "https://foo.bar.baz/johndoe/repo",
				Path:     "/home/someuser/.cache/kickoff/repositories/foo.bar.baz/johndoe/repo@master",
				Revision: "master",
			},
		},
		{
			name: "url with revision",
			s:    "https://foo.bar.baz/johndoe/repo?revision=feature/foo/bar",
			expected: &RepoRef{
				URL:      "https://foo.bar.baz/johndoe/repo",
				Path:     "/home/someuser/.cache/kickoff/repositories/foo.bar.baz/johndoe/repo@feature%2Ffoo%2Fbar",
				Revision: "feature/foo/bar",
			},
		},
		{
			name: "git url",
			s:    "git://git@github.com/martinohmann/kickoff.git",
			expected: &RepoRef{
				URL:      "git://git@github.com/martinohmann/kickoff.git",
				Path:     "/home/someuser/.cache/kickoff/repositories/github.com/martinohmann/kickoff.git@master",
				Revision: "master",
			},
		},
		{
			name: "invalid url",
			s:    "inval\\:",
			err:  errors.New(`invalid repo URL "inval\\:": parse "inval\\:": first path segment in URL cannot contain colon`),
		},
		{
			name: "invalid query",
			s:    "https://foo.bar.baz?revision=%%",
			err:  errors.New(`invalid URL query "revision=%%": invalid URL escape "%%"`),
		},
		{
			name: "parses homedir paths",
			s:    "~/repo",
			expected: &RepoRef{
				Path: filepath.Join(os.Getenv("HOME"), "repo"),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			ref, err := ParseRepoRef(tc.s)
			if tc.err == nil {
				require.NoError(t, err)
				require.Equal(t, tc.expected, ref)
				require.NoError(t, ref.Validate())

				// Roundtrip
				ref2, err := ParseRepoRef(ref.String())
				require.NoError(t, err)
				require.Equal(t, tc.expected, ref2)
				require.NoError(t, ref.Validate())
			} else {
				require.Error(t, err)
				require.EqualError(t, err, tc.err.Error())
			}
		})
	}
}

func TestRepoRef_Validate(t *testing.T) {
	testCases := []validatorTestCase{
		{
			name: "empty location is invalid",
			v:    &RepoRef{},
			err:  newRepositoryRefError("URL or Path must be set"),
		},
		{
			name: "ref with URL is valid",
			v:    &RepoRef{URL: DefaultRepositoryURL},
		},
		{
			name: "ref with local path is valid",
			v:    &RepoRef{Path: "/tmp"},
		},
		{
			name: "ref with URL and local path is valid",
			v: &RepoRef{
				URL:  DefaultRepositoryURL,
				Path: "/tmp",
			},
		},
		{
			name: "ref with URL and revision is valid",
			v: &RepoRef{
				URL:      DefaultRepositoryURL,
				Revision: "master",
			},
		},
		{
			name: "ref with invalid URL",
			v:    &RepoRef{URL: "inval\\:"},
			err:  newRepositoryRefError(`invalid URL: parse "inval\\:": first path segment in URL cannot contain colon`),
		},
	}

	runValidatorTests(t, testCases)
}
