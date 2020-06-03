package repository

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	require := require.New(t)
	tmpdir, err := ioutil.TempDir("", "kickoff-repo-*")
	require.NoError(err)
	defer os.RemoveAll(tmpdir)

	require.NoError(Create(tmpdir, "default"))

	require.DirExists(filepath.Join(tmpdir, "skeletons"))
	require.DirExists(filepath.Join(tmpdir, "skeletons", "default"))
}
