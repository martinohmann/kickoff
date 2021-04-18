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
