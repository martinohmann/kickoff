package skeleton

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/spf13/cobra"
)

func TestCreateCmd_Execute_EmptyOutputDir(t *testing.T) {
	cmd := newCreateCmd()
	cmd.SetArgs([]string{""})

	err := cmd.Execute()
	if err != cmdutil.ErrEmptyOutputDir {
		t.Fatalf("expected error %v, got %v", cmdutil.ErrEmptyOutputDir, err)
	}
}

func TestCreateCmd_Execute_DirExists(t *testing.T) {
	cmd := newCreateCmd()
	cmd.SetArgs([]string{"."})

	dir, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}

	expectedErr := fmt.Errorf("output dir %s already exists, add --force to overwrite", dir)

	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if expectedErr.Error() != err.Error() {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestCreateCmd_Execute(t *testing.T) {
	name, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(name)

	outputDir := filepath.Join(name, "myskeleton")

	cmd := newCreateCmd()
	cmd.SetArgs([]string{outputDir})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestCreateCmd_Execute_Force(t *testing.T) {
	name, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(name)

	outputDir := filepath.Join(name, "myskeleton")

	cmd := newCreateCmd()
	cmd.SetArgs([]string{outputDir})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	cmd = newCreateCmd()
	cmd.SetArgs([]string{outputDir, "--force"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func newCreateCmd() *cobra.Command {
	streams, _, _, _ := cli.NewTestIOStreams()
	return NewCreateCmd(streams)
}
