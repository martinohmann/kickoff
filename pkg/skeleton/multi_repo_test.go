package skeleton

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultiRepo_EmptyRepos_Error(t *testing.T) {
	_, err := NewMultiRepo(nil)
	assert.Equal(t, ErrNoRepositories, err)
}

func TestMultiRepo_EmptyRepoName_Error(t *testing.T) {
	_, err := NewMultiRepo(map[string]string{
		"default": "path/to/foo",
		"":        "path/to/bar",
	})

	expectedErr := errors.New("repository with url path/to/bar was configured with an empty name, please fix your config")
	assert.Equal(t, expectedErr, err)
}

func TestMultiRepo_SkeletonInfo(t *testing.T) {
	pwd, _ := os.Getwd()

	tests := []struct {
		name        string
		skeleton    string
		repos       map[string]string
		expected    *Info
		expectedErr error
	}{
		{
			name:        "skeleton not present in any repo",
			skeleton:    "nonexistent",
			expectedErr: errors.New(`skeleton "nonexistent" not found`),
		},
		{
			name:        "skeleton exists in multiple repos",
			skeleton:    "minimal",
			expectedErr: errors.New(`skeleton "minimal" found in multiple repositories: default, default-copy. explicitly provide <repo-name>:<skeleton-name> to select one`),
		},
		{
			name:     "ambiguous skeleton can be prefixed with repo name to fetch it",
			skeleton: "default-copy:minimal",
			expected: &Info{
				Name: "minimal",
				Path: filepath.Join(pwd, "../testdata/repos/repo2/minimal"),
				Repo: &RepositoryInfo{
					Local: true,
					Name:  "default-copy",
					Path:  filepath.Join(pwd, "../testdata/repos/repo2"),
				},
			},
		},
		{
			name:     "skeletons present in only one repo can be fetched without prefix",
			skeleton: "simple",
			expected: &Info{
				Name: "simple",
				Path: filepath.Join(pwd, "../testdata/repos/repo3/simple"),
				Repo: &RepositoryInfo{
					Local: true,
					Name:  "other",
					Path:  filepath.Join(pwd, "../testdata/repos/repo3"),
				},
			},
		},
		{
			name:     "fails if a repository returned an error #1",
			skeleton: "minimal",
			repos: map[string]string{
				"default": "../testdata/repos/repo1",
				"broken":  "../testdata/repos/nonexistent",
			},
			expectedErr: fmt.Errorf(`stat %s: no such file or directory`, filepath.Join(pwd, "../testdata/repos/nonexistent")),
		},
		{
			name:     "fails if a named repository returned an error #2",
			skeleton: "broken:a-skeleton",
			repos: map[string]string{
				"default": "../testdata/repos/repo1",
				"broken":  "../testdata/repos/nonexistent",
			},
			expectedErr: fmt.Errorf(`stat %s: no such file or directory`, filepath.Join(pwd, "../testdata/repos/nonexistent")),
		},
		{
			name:        "unknown repo name",
			skeleton:    "unknown-repo:someskeleton",
			expectedErr: errors.New(`no skeleton repository configured with name "unknown-repo"`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repos := test.repos
			if repos == nil {
				repos = map[string]string{
					"default":      "../testdata/repos/repo1",
					"default-copy": "../testdata/repos/repo2",
					"other":        "../testdata/repos/repo3",
				}
			}

			repo, err := NewMultiRepo(repos)
			require.NoError(t, err)

			info, err := repo.SkeletonInfo(test.skeleton)

			if test.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, test.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected, info)
			}
		})
	}
}

func TestMultiRepo_SkeletonInfos(t *testing.T) {
	pwd, _ := os.Getwd()

	tests := []struct {
		name        string
		repos       map[string]string
		expected    []*Info
		expectedErr error
	}{
		{
			name: "fails if a there is a nonexistent repo configured",
			repos: map[string]string{
				"default": "../testdata/repos/repo1",
				"broken":  "../testdata/repos/nonexistent",
			},
			expectedErr: fmt.Errorf(`stat %s: no such file or directory`, filepath.Join(pwd, "../testdata/repos/nonexistent")),
		},
		{
			name: "lists all available skeletons",
			expected: []*Info{
				{
					Name: "advanced",
					Path: filepath.Join(pwd, "../testdata/repos/repo1/advanced"),
					Repo: &RepositoryInfo{
						Local: true,
						Name:  "default",
						Path:  filepath.Join(pwd, "../testdata/repos/repo1"),
					},
				},
				{
					Name: "minimal",
					Path: filepath.Join(pwd, "../testdata/repos/repo1/minimal"),
					Repo: &RepositoryInfo{
						Local: true,
						Name:  "default",
						Path:  filepath.Join(pwd, "../testdata/repos/repo1"),
					},
				},
				{
					Name: "minimal",
					Path: filepath.Join(pwd, "../testdata/repos/repo2/minimal"),
					Repo: &RepositoryInfo{
						Local: true,
						Name:  "default-copy",
						Path:  filepath.Join(pwd, "../testdata/repos/repo2"),
					},
				},
				{
					Name: "simple",
					Path: filepath.Join(pwd, "../testdata/repos/repo3/simple"),
					Repo: &RepositoryInfo{
						Local: true,
						Name:  "other",
						Path:  filepath.Join(pwd, "../testdata/repos/repo3"),
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repos := test.repos
			if repos == nil {
				repos = map[string]string{
					"default":      "../testdata/repos/repo1",
					"default-copy": "../testdata/repos/repo2",
					"other":        "../testdata/repos/repo3",
				}
			}

			repo, err := NewMultiRepo(repos)
			require.NoError(t, err)

			infos, err := repo.SkeletonInfos()

			if test.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, test.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected, infos)
			}
		})
	}
}
