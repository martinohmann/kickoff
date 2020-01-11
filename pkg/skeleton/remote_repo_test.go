// +build integration

package skeleton

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenRepository_RemoteRepo(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "kickoff-repos-")
	require.NoError(t, err)

	oldCache := LocalCache
	LocalCache = tmpdir
	defer func() {
		LocalCache = oldCache
		os.RemoveAll(tmpdir)
	}()

	_, err = OpenRepository("https://github.com/martinohmann/kickoff-skeletons?branch=master")
	require.NoError(t, err)

	assert.DirExists(t, filepath.Join(LocalCache, "github.com/martinohmann/kickoff-skeletons/.git"))
}
