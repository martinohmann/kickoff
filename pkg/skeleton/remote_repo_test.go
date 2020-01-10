// +build integration

package skeleton

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestOpenRepository_RemoteRepo(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "kickoff-repos-")
	if err != nil {
		t.Fatal(err)
	}
	oldCache := LocalCache
	LocalCache = tmpdir
	defer func() {
		LocalCache = oldCache
		os.RemoveAll(tmpdir)
	}()

	_, err = OpenRepository("https://github.com/martinohmann/kickoff-skeletons?branch=master")
	if err != nil {
		t.Fatalf("expected nil error but got: %v", err)
	}
}
