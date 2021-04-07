package repository

import (
	"context"
	"testing"

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

	assert.Equal(t, "advanced", skeleton.Info.Name)
	assert.Equal(t, "the-repo", skeleton.Info.Repo.Name)
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
	require := require.New(t)

	repo, err := NewNamed("the-repo", "../testdata/repos/advanced")
	require.NoError(err)

	t.Run("it recursively loads skeletons with its parents", func(t *testing.T) {
		skeleton, err := LoadSkeleton(context.Background(), repo, "childofchild")
		require.NoError(err)
		require.NotNil(skeleton.Parent)
		require.Equal("child", skeleton.Parent.Info.Name)
		require.NotNil(skeleton.Parent.Parent)
		require.Equal("parent", skeleton.Parent.Parent.Info.Name)
		require.Nil(skeleton.Parent.Parent.Parent)
	})

	t.Run("it detects dependency cycles while loading parent skeletons", func(t *testing.T) {
		_, err := LoadSkeleton(context.Background(), repo, "cyclea")
		require.Error(err)
		require.EqualError(err, `failed to load skeleton: dependency cycle detected for parent: kickoff.ParentRef{SkeletonName:"cycleb", RepositoryURL:"../.."}`)
	})
}
