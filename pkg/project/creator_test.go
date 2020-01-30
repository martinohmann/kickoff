package project

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/martinohmann/kickoff/pkg/file"
	"github.com/martinohmann/kickoff/pkg/skeleton"
	"github.com/martinohmann/kickoff/pkg/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	tmpdir, remove := newTempDir(t)
	defer remove()

	loader, err := skeleton.NewSingleRepositoryLoader("../testdata/repos/repo1")
	if err != nil {
		t.Fatal(err)
	}

	skeleton, err := loader.LoadSkeleton("advanced")
	if err != nil {
		t.Fatal(err)
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
	tmpdir, remove := newTempDir(t)
	defer remove()

	loader, err := skeleton.NewSingleRepositoryLoader("../testdata/repos/repo1")
	if err != nil {
		t.Fatal(err)
	}

	skeleton, err := loader.LoadSkeleton("advanced")
	if err != nil {
		t.Fatal(err)
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
	tmpdir, remove := newTempDir(t)
	defer remove()

	loader, err := skeleton.NewSingleRepositoryLoader("../testdata/repos/repo1")
	if err != nil {
		t.Fatal(err)
	}

	skeleton, err := loader.LoadSkeleton("advanced")
	if err != nil {
		t.Fatal(err)
	}

	err = Create(skeleton, tmpdir, &CreateOptions{
		DryRun: true,
		Values: template.Values{
			"filename": "../../",
		},
	})
	require.Error(t, err)

	expectedErr := errors.New(`templated filename "{{.Values.filename}}" injected illegal directory traversal: ../../`)

	assert.Equal(t, expectedErr, err)
}

func TestCreate_EmptyTemplateFilename(t *testing.T) {
	tmpdir, remove := newTempDir(t)
	defer remove()

	loader, err := skeleton.NewSingleRepositoryLoader("../testdata/repos/repo1")
	if err != nil {
		t.Fatal(err)
	}

	skeleton, err := loader.LoadSkeleton("advanced")
	if err != nil {
		t.Fatal(err)
	}

	err = Create(skeleton, tmpdir, &CreateOptions{
		DryRun: true,
		Values: template.Values{
			"filename": "",
		},
	})
	require.Error(t, err)

	expectedErr := errors.New(`templated filename "{{.Values.filename}}" resolved to an empty string`)

	assert.Equal(t, expectedErr, err)
}

func newTempDir(t *testing.T) (string, func()) {
	tmpdir, err := ioutil.TempDir("", "kickoff-")
	if err != nil {
		t.Fatal(err)
	}

	return tmpdir, func() { os.RemoveAll(tmpdir) }
}
