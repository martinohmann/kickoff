package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadSkeletons(t *testing.T) {
	repo, err := NewNamed("the-repo", "../testdata/repos/repo1")
	require.NoError(t, err)

	skeletons, err := LoadSkeletons(context.Background(), repo, []string{"advanced"})
	require.NoError(t, err)

	require.Len(t, skeletons, 1)

	skeleton := skeletons[0]

	assert.Equal(t, "advanced", skeleton.Ref.Name)
	assert.Equal(t, "the-repo", skeleton.Ref.Repo.Name)
	assert.Len(t, skeleton.Files, 4)
	assert.Equal(t,
		template.Values{
			"somekey":         "somevalue",
			"filename":        "foobar",
			"optionalContent": "",
			"travis": map[string]interface{}{
				"enabled": false,
			},
		},
		skeleton.Values,
	)
}

func TestLoadSkeleton(t *testing.T) {
	repo, err := NewNamed("the-repo", "../testdata/repos/advanced")
	require.NoError(t, err)

	t.Run("it recursively loads skeletons with its parents", func(t *testing.T) {
		require := require.New(t)

		skeleton, err := LoadSkeleton(context.Background(), repo, "childofchild")
		require.NoError(err)
		require.NotNil(skeleton.Parent)
		require.Equal("child", skeleton.Parent.Ref.Name)
		require.NotNil(skeleton.Parent.Parent)
		require.Equal("parent", skeleton.Parent.Parent.Ref.Name)
		require.Nil(skeleton.Parent.Parent.Parent)
	})

	t.Run("it detects dependency cycles while loading parent skeletons", func(t *testing.T) {
		_, err := LoadSkeleton(context.Background(), repo, "cyclea")
		require.Error(t, err)
		require.EqualError(t, err, `failed to load skeleton: dependency cycle detected for parent: kickoff.ParentRef{SkeletonName:"cycleb", RepositoryURL:""}`)
	})
}

func TestGetParentRepoRef(t *testing.T) {
	testCases := []struct {
		name        string
		repoRef     *kickoff.RepoRef
		parentRef   kickoff.ParentRef
		expected    *kickoff.RepoRef
		expectedErr error
	}{
		{
			name:      "assume same repo if repositoryURL is empty (local)",
			repoRef:   &kickoff.RepoRef{Name: "foo", Path: "/tmp/foo"},
			parentRef: kickoff.ParentRef{},
			expected:  &kickoff.RepoRef{Name: "foo", Path: "/tmp/foo"},
		},
		{
			name:      "assume same repo if repositoryURL is empty (remote)",
			repoRef:   &kickoff.RepoRef{Name: "foo", URL: "https://foo.bar.baz"},
			parentRef: kickoff.ParentRef{},
			expected:  &kickoff.RepoRef{Name: "foo", URL: "https://foo.bar.baz"},
		},
		{
			name:        "invalid parent ref causes error",
			repoRef:     &kickoff.RepoRef{Name: "foo", URL: "https://foo.bar.baz"},
			parentRef:   kickoff.ParentRef{RepositoryURL: "invalid\\:"},
			expectedErr: errors.New(`invalid repo URL "invalid\\:": parse "invalid\\:": first path segment in URL cannot contain colon`),
		},
		{
			name:      "remote parent repository",
			repoRef:   &kickoff.RepoRef{Name: "foo", Path: "/tmp/foo"},
			parentRef: kickoff.ParentRef{RepositoryURL: "https://foo.bar.baz"},
			expected:  &kickoff.RepoRef{URL: "https://foo.bar.baz", Revision: "master"},
		},
		{
			name:        "cannot reference local repo from remote repo",
			repoRef:     &kickoff.RepoRef{Name: "foo", URL: "https://foo.bar.baz"},
			parentRef:   kickoff.ParentRef{RepositoryURL: "/tmp/foo"},
			expectedErr: errors.New(`cannot reference skeleton from local path "/tmp/foo" as parent in remote repository "https://foo.bar.baz"`),
		},
		{
			name:      "prefixes relative paths with local path",
			repoRef:   &kickoff.RepoRef{Name: "foo", Path: "/tmp/foo"},
			parentRef: kickoff.ParentRef{RepositoryURL: "../baz"},
			expected:  &kickoff.RepoRef{Path: "/tmp/baz"},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ref, err := getParentRepoRef(tc.repoRef, tc.parentRef)
			if tc.expectedErr != nil {
				assert.EqualError(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, ref)
			}
		})
	}
}
