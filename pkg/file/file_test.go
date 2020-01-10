package file

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsDirectory(t *testing.T) {
	ok, err := IsDirectory(".")
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = IsDirectory("file_test.go")
	require.NoError(t, err)
	assert.False(t, ok)

	ok, err = IsDirectory("nonexistent.go")
	require.Error(t, err)
	assert.False(t, ok)
}

func TestExists(t *testing.T) {
	ok := Exists(".")
	assert.True(t, ok)

	ok = Exists("nonexistent.go")
	assert.False(t, ok)
}

func TestCopy(t *testing.T) {
	tmpf, err := ioutil.TempFile("", "kickoff-*")
	require.NoError(t, err)
	filecopy := tmpf.Name() + ".bak"
	defer func() {
		os.Remove(tmpf.Name())
		os.Remove(filecopy)
	}()
	err = tmpf.Close()
	require.NoError(t, err)

	err = ioutil.WriteFile(tmpf.Name(), []byte("foobar"), 0644)
	require.NoError(t, err)

	err = Copy(tmpf.Name(), filecopy)
	require.NoError(t, err)

	tmpfi, err := os.Stat(tmpf.Name())
	require.NoError(t, err)

	fi, err := os.Stat(filecopy)
	require.NoError(t, err)

	assert.Equal(t, tmpfi.Mode(), fi.Mode())
	assert.Equal(t, tmpfi.Size(), fi.Size())
}
