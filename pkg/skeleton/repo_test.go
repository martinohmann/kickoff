package skeleton

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpenRepository_Local(t *testing.T) {
	repo, err := OpenRepository("../testdata/repos/repo1")
	if err != nil {
		t.Fatalf("expected nil error but got: %v", err)
	}

	skel, err := repo.SkeletonInfo("advanced")
	if err != nil {
		t.Fatalf("expected nil error but got: %v", err)
	}

	pwd, _ := os.Getwd()

	expected := &Info{
		Name: "advanced",
		Path: filepath.Join(pwd, "../testdata/repos/repo1/advanced"),
		Repo: &RepositoryInfo{
			Local: true,
			Path:  filepath.Join(pwd, "../testdata/repos/repo1"),
		},
	}

	assert.Equal(t, expected, skel)
}

func TestOpenRepository_LocalError(t *testing.T) {
	_, err := OpenRepository("/nonexistent/local/dir")
	if err == nil {
		t.Fatal("expected error but got nil")
	}
}

func TestFindSkeletons(t *testing.T) {
	skeletons, err := findSkeletons(nil, "../testdata/repos/advanced")
	if err != nil {
		t.Fatalf("expected nil error but got: %v", err)
	}

	pwd, _ := os.Getwd()

	expected := []*Info{
		{Name: "bar", Path: filepath.Join(pwd, "../testdata/repos/advanced/bar")},
		{Name: "child", Path: filepath.Join(pwd, "../testdata/repos/advanced/child")},
		{Name: "childofchild", Path: filepath.Join(pwd, "../testdata/repos/advanced/childofchild")},
		{Name: "cyclea", Path: filepath.Join(pwd, "../testdata/repos/advanced/cyclea")},
		{Name: "cycleb", Path: filepath.Join(pwd, "../testdata/repos/advanced/cycleb")},
		{Name: "cyclec", Path: filepath.Join(pwd, "../testdata/repos/advanced/cyclec")},
		{Name: "foo/bar", Path: filepath.Join(pwd, "../testdata/repos/advanced/foo/bar")},
		{Name: "nested/dir", Path: filepath.Join(pwd, "../testdata/repos/advanced/nested/dir")},
		{Name: "parent", Path: filepath.Join(pwd, "../testdata/repos/advanced/parent")},
	}

	assert.Equal(t, expected, skeletons)
}

func TestFindSkeletons_Error(t *testing.T) {
	_, err := findSkeletons(nil, "nonexistent")
	if err == nil {
		t.Fatalf("expected error but got nil")
	}
}
