package skeleton

import (
	"path/filepath"
	"testing"

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
