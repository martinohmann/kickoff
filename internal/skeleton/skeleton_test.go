package skeleton

import (
	"path/filepath"
	"testing"

	"github.com/martinohmann/kickoff/internal/kickoff"
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

func TestMerge(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	t.Run("merging empty list returns error", func(t *testing.T) {
		_, err := Merge()
		require.Equal(ErrMergeEmpty, err)
	})

	t.Run("merging one returns identity", func(t *testing.T) {
		s0 := &kickoff.Skeleton{}

		s1, err := Merge(s0)
		require.NoError(err)
		assert.Same(s0, s1)
	})

	t.Run("merges skeleton values", func(t *testing.T) {
		s0 := &kickoff.Skeleton{Values: template.Values{"foo": "bar", "baz": false}}
		s1 := &kickoff.Skeleton{Values: template.Values{"qux": 42, "baz": true}}

		s, err := Merge(s0, s1)
		require.NoError(err)
		assert.Equal(template.Values{"foo": "bar", "baz": true, "qux": 42}, s.Values)
	})

	t.Run("merges skeleton files", func(t *testing.T) {
		s0 := &kickoff.Skeleton{
			Files: []kickoff.File{
				&kickoff.FileRef{RelPath: "somefile.txt", AbsPath: "/s0/somefile.txt"},
				&kickoff.FileRef{RelPath: "sometemplate.json.skel", AbsPath: "/s0/sometemplate.json.skel"},
				&kickoff.FileRef{RelPath: "somedir", AbsPath: "/s0/somedir"},
				&kickoff.FileRef{RelPath: "somedir/somefile", AbsPath: "/s0/somedir/somefile"},
			},
		}
		s1 := &kickoff.Skeleton{
			Files: []kickoff.File{
				&kickoff.FileRef{RelPath: "somefile.txt", AbsPath: "/s1/somefile.txt"},
				&kickoff.FileRef{RelPath: "someothertemplate.json.skel", AbsPath: "/s1/someothertemplate.json.skel"},
				&kickoff.FileRef{RelPath: "somedir", AbsPath: "/s1/somedir"},
				&kickoff.FileRef{RelPath: "somedir/someotherfile", AbsPath: "/s1/somedir/someotherfile"},
			},
		}

		s, err := Merge(s0, s1)
		require.NoError(err)

		expectedFiles := []kickoff.File{
			&kickoff.FileRef{RelPath: "somedir", AbsPath: "/s1/somedir"},
			&kickoff.FileRef{RelPath: "somedir/somefile", AbsPath: "/s0/somedir/somefile"},
			&kickoff.FileRef{RelPath: "somedir/someotherfile", AbsPath: "/s1/somedir/someotherfile"},
			&kickoff.FileRef{RelPath: "somefile.txt", AbsPath: "/s1/somefile.txt"},
			&kickoff.FileRef{RelPath: "someothertemplate.json.skel", AbsPath: "/s1/someothertemplate.json.skel"},
			&kickoff.FileRef{RelPath: "sometemplate.json.skel", AbsPath: "/s0/sometemplate.json.skel"},
		}

		assert.Equal(expectedFiles, s.Files)
	})
}
