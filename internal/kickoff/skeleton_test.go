package kickoff

import (
	"testing"

	"github.com/martinohmann/kickoff/internal/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkeleton(t *testing.T) {
	assert := assert.New(t)

	t.Run("string representation", func(t *testing.T) {
		s0 := &Skeleton{Ref: &SkeletonRef{Name: "foo"}}
		assert.Equal("foo", s0.String())

		s1 := &Skeleton{Ref: &SkeletonRef{Name: "bar", Repo: &RepoRef{Name: "repo"}}, Parent: s0}
		assert.Equal("foo->repo:bar", s1.String())

		s2 := &Skeleton{Parent: s1}
		assert.Equal("foo->repo:bar-><anonymous-skeleton>", s2.String())
	})
}

func TestSkeletonRef(t *testing.T) {
	t.Run("string representation", func(t *testing.T) {
		info := &SkeletonRef{Name: "default"}
		assert.Equal(t, "default", info.String())

		info = &SkeletonRef{Name: "default", Repo: &RepoRef{Name: "the-repo"}}
		assert.Equal(t, "the-repo:default", info.String())
	})

	t.Run("load skeleton config from info", func(t *testing.T) {
		info := &SkeletonRef{
			Name: "default",
			Path: "../testdata/repos/repo1/skeletons/minimal",
		}

		config, err := info.LoadConfig()
		require.NoError(t, err)

		expectedConfig := &SkeletonConfig{
			Values: template.Values{"foo": "bar"},
		}

		assert.Equal(t, expectedConfig, config)
	})
}
