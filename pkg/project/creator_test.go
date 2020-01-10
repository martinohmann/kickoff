package project

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/martinohmann/kickoff/pkg/config"
	"github.com/martinohmann/kickoff/pkg/file"
	"github.com/martinohmann/kickoff/pkg/skeleton"
	"github.com/martinohmann/kickoff/pkg/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "kickoff-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	path, err := filepath.Abs("testdata/skeletons/test-skeleton")
	if err != nil {
		t.Fatal(err)
	}

	skeleton := &skeleton.Info{
		Name: "test-skeleton",
		Path: path,
	}

	outputDir := filepath.Join(tmpdir, "myproject")

	err = Create(skeleton, outputDir, &CreateOptions{})
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(outputDir, "README.md"))
	assert.FileExists(t, filepath.Join(outputDir, "foobar", "somefile.yaml"))

	if file.Exists(filepath.Join(outputDir, ".kickoff.yaml")) {
		assert.Fail(t, "expected file %s to be absent but it is not", filepath.Join(outputDir, ".kickoff.yaml"))
	}

	assert.DirExists(t, filepath.Join(outputDir, ".git"))
}

func TestCreate_DryRun(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "kickoff-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	path, err := filepath.Abs("testdata/skeletons/test-skeleton")
	if err != nil {
		t.Fatal(err)
	}

	skeleton := &skeleton.Info{
		Name: "test-skeleton",
		Path: path,
	}

	err = Create(skeleton, tmpdir, &CreateOptions{
		DryRun: true,
	})
	require.NoError(t, err)

	infos, err := ioutil.ReadDir(tmpdir)
	require.NoError(t, err)
	assert.Len(t, infos, 0)
}

func TestCreate_IllegalTemplateFilename(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "kickoff-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	path, err := filepath.Abs("testdata/skeletons/test-skeleton")
	if err != nil {
		t.Fatal(err)
	}

	skeleton := &skeleton.Info{
		Name: "test-skeleton",
		Path: path,
	}

	err = Create(skeleton, tmpdir, &CreateOptions{
		DryRun: true,
		Config: config.Config{
			Values: template.Values{
				"filename": "../../",
			},
		},
	})
	require.Error(t, err)

	expectedErr := errors.New(`templated filename "{{.Values.filename}}" injected illegal directory traversal: ../../`)

	assert.Equal(t, expectedErr, err)
}

func TestCreate_EmptyTemplateFilename(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "kickoff-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	path, err := filepath.Abs("testdata/skeletons/test-skeleton")
	if err != nil {
		t.Fatal(err)
	}

	skeleton := &skeleton.Info{
		Name: "test-skeleton",
		Path: path,
	}

	err = Create(skeleton, tmpdir, &CreateOptions{
		DryRun: true,
		Config: config.Config{
			Values: template.Values{
				"filename": "",
			},
		},
	})
	require.Error(t, err)

	expectedErr := errors.New(`templated filename "{{.Values.filename}}" resolved to an empty string`)

	assert.Equal(t, expectedErr, err)
}
