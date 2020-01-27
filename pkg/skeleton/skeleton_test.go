package skeleton

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestIsInsideSkeletonDir(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expected    bool
		expectedErr error
	}{
		{
			name:     "not inside of skeleton dir",
			path:     "../testdata/repos/advanced/notaskeleton/somefile",
			expected: false,
		},
		{
			name:     "inside of skeleton dir",
			path:     "../testdata/repos/advanced/bar/subdir/somefile.txt",
			expected: true,
		},
		{
			name:     "path is a skeleton dir",
			path:     "../testdata/repos/advanced/bar",
			expected: false,
		},
		{
			name:     "file does not exist",
			path:     "nonexistent",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := IsInsideSkeletonDir(test.path)
			if test.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, test.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func TestFindSkeletonDir(t *testing.T) {
	pwd, _ := os.Getwd()

	tests := []struct {
		name        string
		path        string
		expected    string
		expectedErr error
	}{
		{
			name:        "not inside of skeleton dir",
			path:        "../testdata/repos/advanced/notaskeleton/somefile",
			expectedErr: ErrDirNotFound,
		},
		{
			name:     "skeleton dir",
			path:     "../testdata/repos/advanced/bar",
			expected: filepath.Join(pwd, "../testdata/repos/advanced/bar"),
		},
		{
			name:     "dir inside of skeleton dir",
			path:     "../testdata/repos/advanced/bar/subdir",
			expected: filepath.Join(pwd, "../testdata/repos/advanced/bar"),
		},
		{
			name:     "file in dir inside of skeleton dir",
			path:     "../testdata/repos/advanced/bar/subdir/somefile.txt",
			expected: filepath.Join(pwd, "../testdata/repos/advanced/bar"),
		},
		{
			name:     "file inside of skeleton dir",
			path:     "../testdata/repos/advanced/bar/.kickoff.yaml",
			expected: filepath.Join(pwd, "../testdata/repos/advanced/bar"),
		},
		{
			name:     "nonexistent file inside of skeleton dir",
			path:     "../testdata/repos/advanced/bar/baz/nonexistent.txt",
			expected: filepath.Join(pwd, "../testdata/repos/advanced/bar"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := FindSkeletonDir(test.path)
			if test.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, test.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}
