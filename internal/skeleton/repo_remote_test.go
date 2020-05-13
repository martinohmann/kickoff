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

func TestOpenRepository_Remote(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "kickoff-repos-")
	require.NoError(t, err)

	oldCache := LocalCache
	LocalCache = tmpdir
	defer func() {
		LocalCache = oldCache
		os.RemoveAll(tmpdir)
	}()

	_, err = OpenRepository("https://github.com/martinohmann/kickoff-skeletons?revision=f51e7021f4a8bc344157f33b78a10bc4718eb65a")
	require.NoError(t, err)

	assert.DirExists(t, filepath.Join(LocalCache, "github.com/martinohmann/kickoff-skeletons@f51e7021f4a8bc344157f33b78a10bc4718eb65a/.git"))
}
