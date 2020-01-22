package skeleton

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/martinohmann/kickoff/pkg/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	pwd, _ := os.Getwd()

	info := func(name string) *Info {
		return &Info{
			Name: name,
			Path: filepath.Join(pwd, "testdata/skeletons", name),
		}
	}

	tests := []struct {
		name        string
		info        *Info
		expectedErr error
		validate    func(s *Skeleton, t *testing.T)
	}{
		{
			name: "no parent",
			info: info("parent"),
			validate: func(s *Skeleton, t *testing.T) {
				assert.Nil(t, s.Parent)
				assert.Len(t, s.Files, 5)
				assert.Equal(t, template.Values{
					"foo":      "bar",
					"baz":      "qux",
					"somebool": true,
				}, s.Values)
			},
		},
		{
			name: "child with parent",
			info: info("child"),
			validate: func(s *Skeleton, t *testing.T) {
				assert.NotNil(t, s.Parent)
				assert.Len(t, s.Files, 8)
				assert.Equal(t, template.Values{
					"foo":      "foo",
					"baz":      "qux",
					"somebool": false,
				}, s.Values)
			},
		},
		{
			name: "child of child with parent",
			info: info("childofchild"),
			validate: func(s *Skeleton, t *testing.T) {
				require.NotNil(t, s.Parent)
				assert.NotNil(t, s.Parent.Parent)
				assert.Len(t, s.Files, 9)
				assert.Equal(t, template.Values{
					"foo":      "foo",
					"baz":      "qux",
					"somebool": false,
				}, s.Values)
			},
		},
		{
			name:        "dependency cycle",
			info:        info("cyclea"),
			expectedErr: errors.New(`failed to load skeleton: dependency cycle detected for parent: skeleton.Reference{RepositoryURL:"..", SkeletonName:"cycleb"}`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			skel, err := Load(test.info)
			if test.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, test.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				test.validate(skel, t)
			}
		})
	}
}

func TestSkeleton_WalkFiles(t *testing.T) {
	repo, err := OpenRepository("testdata/skeletons")
	require.NoError(t, err)

	skeleton, err := repo.LoadSkeleton("child")
	require.NoError(t, err)

	actualAbs := make([]string, 0)
	actualRel := make([]string, 0)

	err = skeleton.WalkFiles(func(file *File, err error) error {
		if err == nil {
			actualAbs = append(actualAbs, file.AbsPath)
			actualRel = append(actualRel, file.RelPath)
		}

		return err
	})
	require.NoError(t, err)

	pwd, _ := os.Getwd()

	expectedAbs := []string{
		filepath.Join(pwd, "testdata/skeletons/child"),
		filepath.Join(pwd, "testdata/skeletons/child/dir"),
		filepath.Join(pwd, "testdata/skeletons/child/dir/otherdir"),
		filepath.Join(pwd, "testdata/skeletons/child/dir/otherdir/file.txt"),
		filepath.Join(pwd, "testdata/skeletons/parent/dir/template.txt.skel"),
		filepath.Join(pwd, "testdata/skeletons/child/somefile.txt"),
		filepath.Join(pwd, "testdata/skeletons/child/someotherfile.txt"),
		filepath.Join(pwd, "testdata/skeletons/parent/someparentfile.txt"),
	}

	assert.Equal(t, expectedAbs, actualAbs)

	expectedRel := []string{
		".",
		"dir",
		"dir/otherdir",
		"dir/otherdir/file.txt",
		"dir/template.txt.skel",
		"somefile.txt",
		"someotherfile.txt",
		"someparentfile.txt",
	}

	assert.Equal(t, expectedRel, actualRel)
}
