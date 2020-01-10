package skeleton

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func TestOpenRepository_LocalRepo(t *testing.T) {
	r, err := git.PlainInit("testdata/local-repo", false)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("testdata/local-repo/.git")

	w, err := r.Worktree()
	if err != nil {
		t.Fatal(err)
	}

	_, err = w.Add("b-skeleton")
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

	repo, err := OpenRepository("testdata/local-repo?branch=master")
	if err != nil {
		t.Fatalf("expected nil error but got: %v", err)
	}

	skel, err := repo.Skeleton("b-skeleton")
	if err != nil {
		t.Fatalf("expected nil error but got: %v", err)
	}

	pwd, _ := os.Getwd()

	expected := &Info{
		Name: "b-skeleton",
		Path: filepath.Join(pwd, "testdata/local-repo/b-skeleton"),
	}

	if !reflect.DeepEqual(expected, skel) {
		t.Fatalf("expected skeleton %#v, got %#v", expected, skel)
	}
}

func TestOpenRepository_LocalRepoError(t *testing.T) {
	_, err := OpenRepository("testdata/not-a-local-repo?branch=master")
	if err == nil {
		t.Fatal("expected error but got nil")
	}
}

func TestOpenRepository_LocalDir(t *testing.T) {
	repo, err := OpenRepository("testdata/local-dir")
	if err != nil {
		t.Fatalf("expected nil error but got: %v", err)
	}

	skel, err := repo.Skeleton("a-skeleton")
	if err != nil {
		t.Fatalf("expected nil error but got: %v", err)
	}

	pwd, _ := os.Getwd()

	expected := &Info{
		Name: "a-skeleton",
		Path: filepath.Join(pwd, "testdata/local-dir/a-skeleton"),
	}

	if !reflect.DeepEqual(expected, skel) {
		t.Fatalf("expected skeleton %#v, got %#v", expected, skel)
	}
}

func TestOpenRepository_LocalDirError(t *testing.T) {
	_, err := OpenRepository("testdata/not-a-local-dir")
	if err == nil {
		t.Fatal("expected error but got nil")
	}
}