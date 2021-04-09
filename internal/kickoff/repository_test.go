package kickoff

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRepoRef(t *testing.T) {
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
				Revision: "master",
			},
		},
		{
			name: "url with revision",
			s:    "https://foo.bar.baz/johndoe/repo?revision=feature/foo/bar",
			expected: &RepoRef{
				URL:      "https://foo.bar.baz/johndoe/repo",
				Revision: "feature/foo/bar",
			},
		},
		{
			name: "git url",
			s:    "git://git@github.com/martinohmann/kickoff.git",
			expected: &RepoRef{
				URL:      "git://git@github.com/martinohmann/kickoff.git",
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
			name: "ref with URL and local path is invalid",
			v: &RepoRef{
				URL:  DefaultRepositoryURL,
				Path: "/tmp",
			},
			err: newRepositoryRefError(`URL and Path must not be set at the same time`),
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

func TestRepoRef_LocalPath(t *testing.T) {
	testCases := []struct {
		name     string
		expected string
		ref      *RepoRef
	}{
		{
			name:     "local repository",
			ref:      &RepoRef{Path: "/tmp/foo"},
			expected: "/tmp/foo",
		},
		{
			name:     "remote repository",
			ref:      &RepoRef{URL: "https://github.com/martinohmann/kickoff-skeletons"},
			expected: filepath.Join(LocalRepositoryCacheDir, "4c76fb4fd87cd5b1dca9d94fa35751b06f507109b75bd3a4bc35012ed33cecfb"),
		},
		{
			name:     "remote repository with name",
			ref:      &RepoRef{Name: "default", URL: "https://github.com/martinohmann/kickoff-skeletons"},
			expected: filepath.Join(LocalRepositoryCacheDir, "default-4c76fb4fd87cd5b1dca9d94fa35751b06f507109b75bd3a4bc35012ed33cecfb"),
		},
		{
			name:     "remote repository with revision",
			ref:      &RepoRef{URL: "https://github.com/martinohmann/kickoff-skeletons", Revision: "foo/bar"},
			expected: filepath.Join(LocalRepositoryCacheDir, "4c76fb4fd87cd5b1dca9d94fa35751b06f507109b75bd3a4bc35012ed33cecfb"),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			localPath := tc.ref.LocalPath()
			require.Equal(t, tc.expected, localPath)
		})
	}
}

func TestRepoRef_SkeletonPath(t *testing.T) {
	ref := &RepoRef{Name: "foo", URL: "https://github.com/martinohmann/kickoff-skeletons"}

	assert.Equal(t, ref.SkeletonsPath(), filepath.Join(LocalRepositoryCacheDir, "foo-4c76fb4fd87cd5b1dca9d94fa35751b06f507109b75bd3a4bc35012ed33cecfb", "skeletons"))
	assert.Equal(t, ref.SkeletonPath("bar"), filepath.Join(LocalRepositoryCacheDir, "foo-4c76fb4fd87cd5b1dca9d94fa35751b06f507109b75bd3a4bc35012ed33cecfb", "skeletons", "bar"))
}
