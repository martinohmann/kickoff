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

func TestMultiRepo_Skeleton(t *testing.T) {
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
			skeleton:    "a-skeleton",
			expectedErr: errors.New(`skeleton "a-skeleton" found in multiple repositories: default, default-copy. explicitly provide <repo-name>:<skeleton-name> to select one`),
		},
		{
			name:     "ambiguous skeleton can be prefixed with repo name to fetch it",
			skeleton: "default-copy:a-skeleton",
			expected: &Info{
				Name: "a-skeleton",
				Path: filepath.Join(pwd, "testdata/local-dir-copy/a-skeleton"),
				Repo: &RepositoryInfo{
					Local: true,
					Name:  "default-copy",
					Path:  filepath.Join(pwd, "testdata/local-dir-copy"),
				},
			},
		},
		{
			name:     "skeletons present in only one repo can be fetched without prefix",
			skeleton: "b-skeleton",
			expected: &Info{
				Name: "b-skeleton",
				Path: filepath.Join(pwd, "testdata/local-repo/b-skeleton"),
				Repo: &RepositoryInfo{
					Local: true,
					Name:  "other",
					Path:  filepath.Join(pwd, "testdata/local-repo"),
				},
			},
		},
		{
			name:     "fails if a repository returned and error #1",
			skeleton: "a-skeleton",
			repos: map[string]string{
				"default": "testdata/local-dir",
				"broken":  "testdata/nonexistent",
			},
			expectedErr: fmt.Errorf(`stat %s/testdata/nonexistent: no such file or directory`, pwd),
		},
		{
			name:     "fails if a named repository returned and error #2",
			skeleton: "broken:a-skeleton",
			repos: map[string]string{
				"default": "testdata/local-dir",
				"broken":  "testdata/nonexistent",
			},
			expectedErr: fmt.Errorf(`stat %s/testdata/nonexistent: no such file or directory`, pwd),
		},
		{
			name:        "unknown repo name",
			skeleton:    "unknown-repo:a-skeleton",
			expectedErr: errors.New(`no skeleton repository configured with name "unknown-repo"`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repos := test.repos
			if repos == nil {
				repos = map[string]string{
					"default":      "testdata/local-dir",
					"default-copy": "testdata/local-dir-copy",
					"other":        "testdata/local-repo",
				}
			}

			repo, err := NewMultiRepo(repos)
			require.NoError(t, err)

			info, err := repo.Skeleton(test.skeleton)

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

func TestMultiRepo_Skeletons(t *testing.T) {
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
				"default": "testdata/local-dir",
				"broken":  "testdata/nonexistent",
			},
			expectedErr: fmt.Errorf(`stat %s/testdata/nonexistent: no such file or directory`, pwd),
		},
		{
			name: "lists all available skeletons",
			expected: []*Info{
				{
					Name: "a-skeleton",
					Path: filepath.Join(pwd, "testdata/local-dir/a-skeleton"),
					Repo: &RepositoryInfo{
						Local: true,
						Name:  "default",
						Path:  filepath.Join(pwd, "testdata/local-dir"),
					},
				},
				{
					Name: "a-skeleton",
					Path: filepath.Join(pwd, "testdata/local-dir-copy/a-skeleton"),
					Repo: &RepositoryInfo{
						Local: true,
						Name:  "default-copy",
						Path:  filepath.Join(pwd, "testdata/local-dir-copy"),
					},
				},
				{
					Name: "b-skeleton",
					Path: filepath.Join(pwd, "testdata/local-repo/b-skeleton"),
					Repo: &RepositoryInfo{
						Local: true,
						Name:  "other",
						Path:  filepath.Join(pwd, "testdata/local-repo"),
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
					"default":      "testdata/local-dir",
					"default-copy": "testdata/local-dir-copy",
					"other":        "testdata/local-repo",
				}
			}

			repo, err := NewMultiRepo(repos)
			require.NoError(t, err)

			infos, err := repo.Skeletons()

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
