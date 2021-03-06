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

		s1 := &Skeleton{Ref: &SkeletonRef{Name: "bar", Repo: &RepoRef{Name: "repo"}}}
		assert.Equal("repo:bar", s1.String())
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
			Description: "minimal description",
			Values:      template.Values{"foo": "bar"},
		}

		assert.Equal(t, expectedConfig, config)
	})
}

func TestSkeletonRef_Validate(t *testing.T) {
	testCases := []validatorTestCase{
		{
			name: "empty name is invalid",
			v:    &SkeletonRef{},
			err:  newSkeletonRefError("Name must not be empty"),
		},
		{
			name: "non-empty name is valid",
			v:    &SkeletonRef{Name: "foo"},
		},
		{
			name: "ref with non-nil but empty repo ref is invalid",
			v:    &SkeletonRef{Name: "foo", Repo: &RepoRef{}},
			err:  newRepositoryRefError("URL or Path must be set"),
		},
	}

	runValidatorTests(t, testCases)
}

func TestMergeSkeletons(t *testing.T) {
	t.Run("merging empty list returns error", func(t *testing.T) {
		_, err := MergeSkeletons()
		require.Equal(t, ErrMergeEmpty, err)
	})

	t.Run("merging one returns identity", func(t *testing.T) {
		s0 := &Skeleton{}

		s1, err := MergeSkeletons(s0)
		require.NoError(t, err)
		assert.Same(t, s0, s1)
	})

	t.Run("merges skeleton values", func(t *testing.T) {
		s0 := &Skeleton{Values: template.Values{"foo": "bar", "baz": false}}
		s1 := &Skeleton{Values: template.Values{"qux": 42, "baz": true}}

		s, err := MergeSkeletons(s0, s1)
		require.NoError(t, err)
		assert.Equal(t, template.Values{"foo": "bar", "baz": true, "qux": 42}, s.Values)
	})

	t.Run("merges skeleton files", func(t *testing.T) {
		s0 := &Skeleton{
			Files: []*BufferedFile{
				{RelPath: "somefile.txt", Content: []byte("a")},
				{RelPath: "sometemplate.json.skel"},
				{RelPath: "somedir"},
				{RelPath: "somedir/somefile"},
			},
		}
		s1 := &Skeleton{
			Files: []*BufferedFile{
				{RelPath: "somefile.txt", Content: []byte("b")},
				{RelPath: "someothertemplate.json.skel"},
				{RelPath: "somedir"},
				{RelPath: "somedir/someotherfile"},
			},
		}

		s, err := MergeSkeletons(s0, s1)
		require.NoError(t, err)

		expectedFiles := []*BufferedFile{
			{RelPath: "somedir"},
			{RelPath: "somedir/somefile"},
			{RelPath: "somedir/someotherfile"},
			{RelPath: "somefile.txt", Content: []byte("b")},
			{RelPath: "someothertemplate.json.skel"},
			{RelPath: "sometemplate.json.skel"},
		}

		assert.Equal(t, expectedFiles, s.Files)
	})
}
