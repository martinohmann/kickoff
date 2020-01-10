package skeleton

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestInitCmd_Execute_EmptyOutputDir(t *testing.T) {
	cmd := NewInitCmd()
	cmd.SetArgs([]string{""})

	err := cmd.Execute()
	if err != ErrEmptyOutputDir {
		t.Fatalf("expected error %v, got %v", ErrEmptyOutputDir, err)
	}
}

func TestInitCmd_Execute_DirExists(t *testing.T) {
	cmd := NewInitCmd()
	cmd.SetArgs([]string{"."})

	dir, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}

	expectedErr := fmt.Errorf("output-dir %s already exists, add --force to overwrite", dir)

	err = cmd.Execute()
	if !reflect.DeepEqual(expectedErr, err) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestInitCmd_Execute(t *testing.T) {
	name, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(name)

	outputDir := filepath.Join(name, "myskeleton")

	cmd := NewInitCmd()
	cmd.SetArgs([]string{outputDir})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestInitCmd_Execute_Force(t *testing.T) {
	name, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(name)

	outputDir := filepath.Join(name, "myskeleton")

	cmd := NewInitCmd()
	cmd.SetArgs([]string{outputDir})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	cmd = NewInitCmd()
	cmd.SetArgs([]string{outputDir, "--force"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}