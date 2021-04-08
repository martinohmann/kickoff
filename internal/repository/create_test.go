package repository

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	require := require.New(t)
	tmpdir, err := ioutil.TempDir("", "kickoff-repo-*")
	require.NoError(err)
	defer os.RemoveAll(tmpdir)

	require.NoError(Create(tmpdir, "default"))

	require.DirExists(filepath.Join(tmpdir, kickoff.SkeletonsDir))
	require.DirExists(filepath.Join(tmpdir, kickoff.SkeletonsDir, "default"))
}
