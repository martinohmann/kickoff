package skeleton

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	require := require.New(t)

	dir, err := ioutil.TempDir("", "kickoff-*")
	require.NoError(err)
	defer os.RemoveAll(dir)

	outputPath := filepath.Join(dir, "myskeleton")

	require.NoError(Create(outputPath))
	require.DirExists(outputPath)
	require.FileExists(filepath.Join(outputPath, "README.md.skel"))
	require.FileExists(filepath.Join(outputPath, ConfigFileName))
}
