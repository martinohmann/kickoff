package project

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/martinohmann/kickoff/internal/gitignore"
	"github.com/martinohmann/kickoff/internal/license"
	"github.com/martinohmann/kickoff/internal/repository"
	"github.com/martinohmann/kickoff/internal/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectedErr error
		setup       func(*dirTester)
		validate    func(*dirTester)
	}{
		{
			name:   "empty config is valid",
			config: &Config{},
		},
		{
			name: "absolute paths on overwrite list cause error",
			config: &Config{
				OverwriteFiles: []string{
					"README.md",
					"./relfile",
					"/tmp/somefile",
				},
			},
			expectedErr: errors.New("found illegal absolute path: /tmp/somefile"),
		},
		{
			name: "absolute paths on skip list cause error",
			config: &Config{
				SkipFiles: []string{
					"./relfile",
					"README.md",
					"/tmp/somefile",
				},
			},
			expectedErr: errors.New("found illegal absolute path: /tmp/somefile"),
		},
		{
			name: "create with license and gitignore file",
			config: &Config{
				Owner: "johndoe",
				Gitignore: &gitignore.Template{
					Content: []byte("somegitignorebody"),
				},
				License: &license.Info{
					Body: `some license [fullname] [year]`,
				},
			},
			validate: func(t *dirTester) {
				t.assertFileAbsent(".kickoff.yaml")
				t.assertFileExists("README.md")
				t.assertFileExists(filepath.Join("foobar", "somefile.yaml"))
				t.assertFileContains(".gitignore", `somegitignorebody`)
				t.assertFileContains("LICENSE", `some license johndoe `+strconv.Itoa(time.Now().Year()))
			},
		},
		{
			name: "illegal directory traversals in rendered filenames are detected",
			config: &Config{
				Values: template.Values{"filename": "../../"},
			},
			expectedErr: errors.New(`templated filename "{{.Values.filename}}" injected illegal directory traversal: ../../`),
		},
		{
			name: "rendering empty filename fails",
			config: &Config{
				Values: template.Values{"filename": ""},
			},
			expectedErr: errors.New(`templated filename "{{.Values.filename}}" resolved to an empty string`),
		},
		{
			name: "does not overwrite existing files",
			config: &Config{
				License: &license.Info{Body: `some license [fullname] [year]`},
			},
			setup: func(t *dirTester) {
				t.mustWriteFile("LICENSE", `do not touch`)
				t.mustWriteFile("README.md", `do not touch`)
			},
			validate: func(t *dirTester) {
				t.assertFileContains("LICENSE", `do not touch`)
				t.assertFileContains("README.md", `do not touch`)
			},
		},
		{
			name: "does not overwrite existing files selectively if WithOverwriteFiles is provided",
			config: &Config{
				License: &license.Info{Body: `some license [fullname] [year]`},
				OverwriteFiles: []string{
					"README.md",
					"./foobar/../foobar/somefile.yaml",
				},
			},
			setup: func(t *dirTester) {
				t.mustWriteFile(filepath.Join("foobar", "somefile.yaml"), `do not touch`)
				t.mustWriteFile("LICENSE", `do not touch`)
				t.mustWriteFile("README.md", `please overwrite`)
			},
			validate: func(t *dirTester) {
				t.assertFileContains(filepath.Join("foobar", "somefile.yaml"), "---\nsomekey: {{.Values.somekey}}\n")
				t.assertFileContains("LICENSE", `do not touch`)
				t.assertFileNotContains("README.md", `please overwrite`)
			},
		},
		{
			name: "does overwrite existing files if WithOverwrite option is set",
			config: &Config{
				License:   &license.Info{Body: `some license [fullname] [year]`},
				Overwrite: true,
			},
			setup: func(t *dirTester) {
				t.mustWriteFile("LICENSE", `please overwrite`)
				t.mustWriteFile("README.md", `please overwrite`)
			},
			validate: func(t *dirTester) {
				t.assertFileNotContains("LICENSE", `please overwrite`)
				t.assertFileNotContains("README.md", `please overwrite`)
			},
		},
		{
			name: "skips files selectively if WithSkipFiles is provided",
			config: &Config{
				License: &license.Info{Body: `some license [fullname] [year]`},
				SkipFiles: []string{
					"README.md",
					"./foobar/../foobar",
				},
			},
			validate: func(t *dirTester) {
				t.assertFileAbsent("foobar")
				t.assertFileAbsent(filepath.Join("foobar", "somefile.yaml"))
				t.assertFileAbsent("README.md")
				t.assertFileExists("LICENSE")
			},
		},
		{
			name: "template errors are returned",
			config: &Config{
				Values: template.Values{"travis": "invalid"},
			},
			expectedErr: errors.New(`failed to render template: template: :4:13: executing "" at <.Values.travis.enabled>: can't evaluate field enabled in type interface {}`),
		},
		{
			name: "errors while resolving templated filenames are returned",
			config: &Config{
				Values: template.Values{"filename": func() {}},
			},
			expectedErr: errors.New(`failed to resolve templated filename "{{.Values.filename}}": failed to render template: template: :1:2: executing "" at <{{.Values.filename}}>: can't print {{.Values.filename}} of type func()`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpdir := t.TempDir()

			repo, err := repository.Open(context.Background(), "../testdata/repos/repo1", nil)
			require.NoError(t, err)

			skeleton, err := repo.LoadSkeleton("advanced")
			require.NoError(t, err)

			tester := &dirTester{T: t, dir: tmpdir}

			if test.setup != nil {
				test.setup(tester)
			}

			test.config.Skeleton = skeleton
			test.config.ProjectDir = tmpdir

			err = Create(test.config)
			if test.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			if test.validate != nil {
				test.validate(tester)
			}
		})
	}
}

type dirTester struct {
	*testing.T
	dir string
}

func (t *dirTester) path(file string) string {
	return filepath.Join(t.dir, file)
}

func (t *dirTester) assertFileContains(file, expectedContent string) {
	contents, err := os.ReadFile(t.path(file))
	require.NoError(t, err)
	assert.Equal(t, expectedContent, string(contents))
}

func (t *dirTester) assertFileNotContains(file, expectedContent string) {
	contents, err := os.ReadFile(t.path(file))
	require.NoError(t, err)
	assert.NotEqual(t, expectedContent, string(contents))
}

func (t *dirTester) assertFileExists(file string) {
	assert.FileExists(t, t.path(file))
}

func (t *dirTester) assertFileAbsent(file string) {
	_, err := os.ReadFile(t.path(file))
	assert.True(t, os.IsNotExist(err))
}

func (t *dirTester) mustWriteFile(file, content string) {
	path := t.path(file)

	err := os.MkdirAll(filepath.Dir(path), 0777)
	require.NoError(t, err)

	err = os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
}
