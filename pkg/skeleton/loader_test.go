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

func TestLoader_LoadSkeletons(t *testing.T) {
	tests := []struct {
		name          string
		skeletonNames []string
		expectedErr   error
		validate      func(s []*Skeleton, t *testing.T)
	}{
		{
			name:          "no parent",
			skeletonNames: []string{"parent"},
			validate: func(s []*Skeleton, t *testing.T) {
				require.Len(t, s, 1)
				assert.Nil(t, s[0].Parent)
				assert.Len(t, s[0].Files, 5)
				assert.Equal(t, template.Values{
					"foo":      "bar",
					"baz":      "qux",
					"somebool": true,
				}, s[0].Values)
			},
		},
		{
			name:          "child with parent",
			skeletonNames: []string{"child"},
			validate: func(s []*Skeleton, t *testing.T) {
				require.Len(t, s, 1)
				assert.NotNil(t, s[0].Parent)
				assert.Len(t, s[0].Files, 8)
				assert.Equal(t, template.Values{
					"foo":      "foo",
					"baz":      "qux",
					"somebool": false,
				}, s[0].Values)
			},
		},
		{
			name:          "child of child with parent",
			skeletonNames: []string{"childofchild"},
			validate: func(s []*Skeleton, t *testing.T) {
				require.Len(t, s, 1)
				require.NotNil(t, s[0].Parent)
				assert.NotNil(t, s[0].Parent.Parent)
				assert.Len(t, s[0].Files, 9)
				assert.Equal(t, template.Values{
					"foo":      "foo",
					"baz":      "qux",
					"somebool": false,
				}, s[0].Values)
			},
		},
		{
			name:          "load multiple",
			skeletonNames: []string{"childofchild", "parent", "child"},
			validate: func(s []*Skeleton, t *testing.T) {
				require.Len(t, s, 3)
				assert.Equal(t, "childofchild", s[0].Info.String())
				assert.Equal(t, "parent", s[1].Info.String())
				assert.Equal(t, "child", s[2].Info.String())
			},
		},
		{
			name:          "dependency cycle",
			skeletonNames: []string{"cyclea"},
			expectedErr:   errors.New(`failed to load skeleton: dependency cycle detected for parent: skeleton.Reference{RepositoryURL:"..", SkeletonName:"cycleb"}`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			loader, err := NewSingleRepositoryLoader("../testdata/repos/advanced")
			require.NoError(t, err)

			skeletons, err := loader.LoadSkeletons(test.skeletonNames)
			if test.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, test.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				test.validate(skeletons, t)
			}
		})
	}
}

func TestSkeleton_WalkFiles(t *testing.T) {
	loader, err := NewSingleRepositoryLoader("../testdata/repos/advanced")
	require.NoError(t, err)

	skeleton, err := loader.LoadSkeleton("child")
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
		filepath.Join(pwd, "../testdata/repos/advanced/child"),
		filepath.Join(pwd, "../testdata/repos/advanced/child/dir"),
		filepath.Join(pwd, "../testdata/repos/advanced/child/dir/otherdir"),
		filepath.Join(pwd, "../testdata/repos/advanced/child/dir/otherdir/file.txt"),
		filepath.Join(pwd, "../testdata/repos/advanced/parent/dir/template.txt.skel"),
		filepath.Join(pwd, "../testdata/repos/advanced/child/somefile.txt"),
		filepath.Join(pwd, "../testdata/repos/advanced/child/someotherfile.txt"),
		filepath.Join(pwd, "../testdata/repos/advanced/parent/someparentfile.txt"),
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
