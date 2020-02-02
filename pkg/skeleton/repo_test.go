package skeleton

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenRepository_Local(t *testing.T) {
	repo, err := OpenRepository("../testdata/repos/repo1")
	require.NoError(t, err)

	skel, err := repo.SkeletonInfo("advanced")
	require.NoError(t, err)

	pwd, _ := os.Getwd()

	expected := &Info{
		Name: "advanced",
		Path: filepath.Join(pwd, "../testdata/repos/repo1/skeletons/advanced"),
		Repo: &RepositoryInfo{
			Local: true,
			Path:  filepath.Join(pwd, "../testdata/repos/repo1"),
		},
	}

	assert.Equal(t, expected, skel)
}

func TestOpenRepository_LocalError(t *testing.T) {
	_, err := OpenRepository("/nonexistent/local/dir")
	require.Error(t, err)
}

func TestOpenRepository_LocalError_InvalidRepo(t *testing.T) {
	pwd, _ := os.Getwd()

	repoPath := filepath.Join(pwd, "../testdata/repos/invalidrepo")

	_, err := OpenRepository(repoPath)
	require.Error(t, err)
	assert.Equal(t, fmt.Sprintf(`%s is not a valid skeleton repository, skeletons/ is not a directory`, repoPath), err.Error())
}

func TestFindSkeletons(t *testing.T) {
	skeletons, err := findSkeletons(nil, "../testdata/repos/advanced/skeletons")
	require.NoError(t, err)

	pwd, _ := os.Getwd()

	expected := []*Info{
		{Name: "bar", Path: filepath.Join(pwd, "../testdata/repos/advanced/skeletons/bar")},
		{Name: "child", Path: filepath.Join(pwd, "../testdata/repos/advanced/skeletons/child")},
		{Name: "childofchild", Path: filepath.Join(pwd, "../testdata/repos/advanced/skeletons/childofchild")},
		{Name: "cyclea", Path: filepath.Join(pwd, "../testdata/repos/advanced/skeletons/cyclea")},
		{Name: "cycleb", Path: filepath.Join(pwd, "../testdata/repos/advanced/skeletons/cycleb")},
		{Name: "cyclec", Path: filepath.Join(pwd, "../testdata/repos/advanced/skeletons/cyclec")},
		{Name: "foo/bar", Path: filepath.Join(pwd, "../testdata/repos/advanced/skeletons/foo/bar")},
		{Name: "nested/dir", Path: filepath.Join(pwd, "../testdata/repos/advanced/skeletons/nested/dir")},
		{Name: "parent", Path: filepath.Join(pwd, "../testdata/repos/advanced/skeletons/parent")},
	}

	assert.Equal(t, expected, skeletons)
}

func TestFindSkeletons_Error(t *testing.T) {
	_, err := findSkeletons(nil, "nonexistent")
	require.Error(t, err)
}
