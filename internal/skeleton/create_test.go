package skeleton

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateRepository(t *testing.T) {
	dir, err := ioutil.TempDir("", "kickoff-*")
	require.NoError(t, err)

	defer os.RemoveAll(dir)

	outputPath := filepath.Join(dir, "repo")

	err = CreateRepository(outputPath, "myskeleton")
	require.NoError(t, err)

	skeletonsDir := filepath.Join(outputPath, "skeletons")
	skeletonDir := filepath.Join(skeletonsDir, "myskeleton")

	assert.DirExists(t, outputPath)
	assert.DirExists(t, skeletonsDir)
	assert.DirExists(t, skeletonDir)
	assert.FileExists(t, filepath.Join(skeletonDir, "README.md.skel"))
	assert.FileExists(t, filepath.Join(skeletonDir, ConfigFileName))
}
