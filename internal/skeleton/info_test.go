package skeleton

import (
	"testing"

	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInfo(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	t.Run("string representation", func(t *testing.T) {
		info := &Info{Name: "default"}
		assert.Equal("default", info.String())

		info = &Info{Name: "default", Repo: &kickoff.RepoRef{Name: "the-repo"}}
		assert.Equal("the-repo:default", info.String())
	})

	t.Run("load skeleton config from info", func(t *testing.T) {
		info := &Info{
			Name: "default",
			Path: "../testdata/repos/repo1/skeletons/minimal",
		}

		config, err := info.LoadConfig()
		require.NoError(err)

		expectedConfig := &kickoff.SkeletonConfig{
			Values: template.Values{"foo": "bar"},
		}

		assert.Equal(expectedConfig, config)
	})
}
