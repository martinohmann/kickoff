package skeleton

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindSkeletons(t *testing.T) {
	skeletons, err := findSkeletons(nil, "testdata/skeletons")
	if err != nil {
		t.Fatalf("expected nil error but got: %v", err)
	}

	pwd, _ := os.Getwd()

	expected := []*Info{
		{Name: "bar", Path: filepath.Join(pwd, "testdata/skeletons/bar")},
		{Name: "foo/bar", Path: filepath.Join(pwd, "testdata/skeletons/foo/bar")},
		{Name: "nested", Path: filepath.Join(pwd, "testdata/skeletons/nested")},
		{Name: "nested/skeletons", Path: filepath.Join(pwd, "testdata/skeletons/nested/skeletons")},
	}

	assert.Equal(t, expected, skeletons)
}

func TestFindSkeletons_Error(t *testing.T) {
	_, err := findSkeletons(nil, "testdata/nonexistent")
	if err == nil {
		t.Fatalf("expected error but got nil")
	}
}
