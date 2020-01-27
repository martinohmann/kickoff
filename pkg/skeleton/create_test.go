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

	assert.DirExists(t, outputPath)
	assert.DirExists(t, filepath.Join(outputPath, "myskeleton"))
	assert.FileExists(t, filepath.Join(outputPath, "myskeleton", "README.md.skel"))
	assert.FileExists(t, filepath.Join(outputPath, "myskeleton", ConfigFileName))
}
