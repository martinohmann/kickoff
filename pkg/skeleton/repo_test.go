package skeleton

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func TestOpenRepository_LocalRepo(t *testing.T) {
	r, err := git.PlainInit("../testdata/repos/repo3", false)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("../testdata/repos/repo3/.git")

	w, err := r.Worktree()
	if err != nil {
		t.Fatal(err)
	}

	_, err = w.Add("simple")
	if err != nil {
		t.Fatal(err)
	}

	_, err = w.Commit("initial", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "John Doe",
			Email: "john@doe.org",
			When:  time.Now(),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	repo, err := OpenRepository("../testdata/repos/repo3?branch=master")
	if err != nil {
		t.Fatalf("expected nil error but got: %v", err)
	}

	skel, err := repo.SkeletonInfo("simple")
	if err != nil {
		t.Fatalf("expected nil error but got: %v", err)
	}

	pwd, _ := os.Getwd()

	expected := &Info{
		Name: "simple",
		Path: filepath.Join(pwd, "../testdata/repos/repo3/simple"),
		Repo: &RepositoryInfo{
			Local:  true,
			Path:   filepath.Join(pwd, "../testdata/repos/repo3"),
			Branch: "master",
		},
	}

	assert.Equal(t, expected, skel)
}

func TestOpenRepository_LocalRepoError(t *testing.T) {
	_, err := OpenRepository("/nonexistent/local/repo?branch=master")
	if err == nil {
		t.Fatal("expected error but got nil")
	}
}

func TestOpenRepository_LocalDir(t *testing.T) {
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

func TestOpenRepository_LocalDirError(t *testing.T) {
	_, err := OpenRepository("/nonexistent/local/dir")
	if err == nil {
		t.Fatal("expected error but got nil")
	}
}
