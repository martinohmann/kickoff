package kickoff

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fileTestCase struct {
	name       string
	mode       os.FileMode
	path       string
	content    []byte
	isTemplate bool
	isDir      bool
}

type fileTestFactory func(t *testing.T, relPath string, content []byte, mode os.FileMode) File

func runFileTests(t *testing.T, factory fileTestFactory) {
	testCases := []fileTestCase{
		{
			name:    "simple",
			path:    "some/path",
			content: []byte(`the content`),
			mode:    0644,
		},
		{
			name:  "dir",
			path:  "some/path",
			mode:  os.ModeDir,
			isDir: true,
		},
		{
			name:       "template",
			path:       "some/path.skel",
			content:    []byte(`the content`),
			mode:       0644,
			isTemplate: true,
		},
		{
			name:       "dir with .skel extention",
			path:       "some/path.skel",
			content:    []byte(`the content`),
			mode:       0755 | os.ModeDir,
			isTemplate: false,
			isDir:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			f := factory(t, tc.path, tc.content, tc.mode)

			assert.Equal(t, tc.path, f.Path())
			assert.Equal(t, tc.mode, f.Mode())

			r, err := f.Reader()
			require.NoError(t, err)

			buf, err := ioutil.ReadAll(r)
			require.NoError(t, err)

			assert.Equal(t, string(tc.content), string(buf))
			assert.Equal(t, tc.isTemplate, f.IsTemplate())
			assert.Equal(t, tc.isDir, f.Mode().IsDir())
		})
	}
}

func TestBufferedFile(t *testing.T) {
	runFileTests(t, func(t *testing.T, relPath string, content []byte, mode os.FileMode) File {
		return NewBufferedFile(relPath, content, mode)
	})
}

func TestFileRef(t *testing.T) {
	runFileTests(t, func(t *testing.T, relPath string, content []byte, mode os.FileMode) File {
		f, err := ioutil.TempFile(t.TempDir(), "file-")
		require.NoError(t, err)
		err = ioutil.WriteFile(f.Name(), content, mode)
		require.NoError(t, err)

		return &FileRef{RelPath: relPath, AbsPath: f.Name(), FileMode: mode}
	})
}
