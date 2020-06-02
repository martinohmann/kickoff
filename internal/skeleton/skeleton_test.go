package skeleton

import (
	"path/filepath"
	"testing"

	"github.com/martinohmann/kickoff/internal/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsInsideSkeletonDir(t *testing.T) {
	assert := assert.New(t)

	t.Run("not inside skeleton dir", func(t *testing.T) {
		ok, _ := IsInsideSkeletonDir("../testdata/repos/advanced/skeletons/notaskeleton/somefile")
		assert.False(ok)
	})

	t.Run("inside skeleton dir", func(t *testing.T) {
		ok, _ := IsInsideSkeletonDir("../testdata/repos/advanced/skeletons/bar/subdir/somefile.txt")
		assert.True(ok)
	})

	t.Run("path is skeleton dir", func(t *testing.T) {
		ok, _ := IsInsideSkeletonDir("../testdata/repos/advanced/skeletons/bar")
		assert.False(ok)
	})

	t.Run("file does not exist", func(t *testing.T) {
		ok, _ := IsInsideSkeletonDir("nonexistent")
		assert.False(ok)
	})
}

func TestFindSkeletonDir(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	abspath := func(path string) string {
		abs, err := filepath.Abs(path)
		require.NoError(err)
		return abs
	}

	t.Run("not inside skeleton dir", func(t *testing.T) {
		_, err := FindSkeletonDir("../testdata/repos/advanced/skeletons/notaskeleton/somefile")
		assert.Equal(ErrDirNotFound, err)
	})

	t.Run("skeleton dir", func(t *testing.T) {
		path, err := FindSkeletonDir("../testdata/repos/advanced/skeletons/bar")
		require.NoError(err)
		assert.Equal(abspath("../testdata/repos/advanced/skeletons/bar"), path)
	})

	t.Run("file in dir inside of skeleton dir", func(t *testing.T) {
		path, err := FindSkeletonDir("../testdata/repos/advanced/skeletons/bar/subdir/somefile.txt")
		require.NoError(err)
		assert.Equal(abspath("../testdata/repos/advanced/skeletons/bar"), path)
	})

	t.Run("nonexistent file inside of skeleton dir", func(t *testing.T) {
		path, err := FindSkeletonDir("../testdata/repos/advanced/skeletons/bar/baz/nonexistent.txt")
		require.NoError(err)
		assert.Equal(abspath("../testdata/repos/advanced/skeletons/bar"), path)
	})
}

func TestSkeleton(t *testing.T) {
	assert := assert.New(t)

	t.Run("string representation", func(t *testing.T) {
		s0 := &Skeleton{Info: &Info{Name: "foo"}}
		assert.Equal("foo", s0.String())

		s1 := &Skeleton{Info: &Info{Name: "bar", Repo: &RepoInfo{Name: "repo"}}, Parent: s0}
		assert.Equal("foo->repo:bar", s1.String())

		s2 := &Skeleton{Parent: s1}
		assert.Equal("foo->repo:bar-><anonymous-skeleton>", s2.String())
	})
}

func TestMerge(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	t.Run("merging empty list returns error", func(t *testing.T) {
		_, err := Merge()
		require.Equal(ErrMergeEmpty, err)
	})

	t.Run("merging one returns identity", func(t *testing.T) {
		s0 := &Skeleton{}

		s1, err := Merge(s0)
		require.NoError(err)
		assert.Same(s0, s1)
	})

	t.Run("merges skeleton values", func(t *testing.T) {
		s0 := &Skeleton{Values: template.Values{"foo": "bar", "baz": false}}
		s1 := &Skeleton{Values: template.Values{"qux": 42, "baz": true}}

		s, err := Merge(s0, s1)
		require.NoError(err)
		assert.Equal(template.Values{"foo": "bar", "baz": true, "qux": 42}, s.Values)
	})

	t.Run("merges skeleton files", func(t *testing.T) {
		s0 := &Skeleton{
			Files: []*File{
				{RelPath: "somefile.txt", AbsPath: "/s0/somefile.txt"},
				{RelPath: "sometemplate.json.skel", AbsPath: "/s0/sometemplate.json.skel"},
				{RelPath: "somedir", AbsPath: "/s0/somedir"},
				{RelPath: "somedir/somefile", AbsPath: "/s0/somedir/somefile"},
			},
		}
		s1 := &Skeleton{
			Files: []*File{
				{RelPath: "somefile.txt", AbsPath: "/s1/somefile.txt"},
				{RelPath: "someothertemplate.json.skel", AbsPath: "/s1/someothertemplate.json.skel"},
				{RelPath: "somedir", AbsPath: "/s1/somedir"},
				{RelPath: "somedir/someotherfile", AbsPath: "/s1/somedir/someotherfile"},
			},
		}

		s, err := Merge(s0, s1)
		require.NoError(err)

		expectedFiles := []*File{
			{RelPath: "somedir", AbsPath: "/s1/somedir"},
			{RelPath: "somedir/somefile", AbsPath: "/s0/somedir/somefile"},
			{RelPath: "somedir/someotherfile", AbsPath: "/s1/somedir/someotherfile"},
			{RelPath: "somefile.txt", AbsPath: "/s1/somefile.txt"},
			{RelPath: "someothertemplate.json.skel", AbsPath: "/s1/someothertemplate.json.skel"},
			{RelPath: "sometemplate.json.skel", AbsPath: "/s0/sometemplate.json.skel"},
		}

		assert.Equal(expectedFiles, s.Files)
	})
}
