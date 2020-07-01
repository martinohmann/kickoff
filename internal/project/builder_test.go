package project

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/discard"
	"github.com/martinohmann/kickoff/internal/config"
	"github.com/martinohmann/kickoff/internal/license"
	"github.com/martinohmann/kickoff/internal/repository"
	"github.com/martinohmann/kickoff/internal/template"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type badFile string

func (f badFile) Path() string               { return string(f) }
func (f badFile) Mode() os.FileMode          { return 0644 }
func (f badFile) Reader() (io.Reader, error) { return nil, errors.New(string(f)) }

func TestBuilder_Build(t *testing.T) {
	log.SetHandler(discard.Default)

	tests := []struct {
		name        string
		config      config.Project
		expectedErr error
		setup       func(t *dirTester, b *Builder)
		validate    func(t *dirTester, b *Builder)
	}{
		{
			name: "create with license and gitignore file",
			config: config.Project{
				Owner: "johndoe",
			},
			setup: func(t *dirTester, b *Builder) {
				b.WithGitignore("somegitignorebody").
					WithLicense(&license.Info{
						Body: `some license [fullname] [year]`,
					})
			},
			validate: func(t *dirTester, b *Builder) {
				t.assertFileAbsent(".kickoff.yaml")
				t.assertFileExists("README.md")
				t.assertFileExists(filepath.Join("foobar", "somefile.yaml"))
				t.assertFileContains(".gitignore", `somegitignorebody`)
				t.assertFileContains("LICENSE", `some license johndoe `+strconv.Itoa(time.Now().Year()))
			},
		},
		{
			name: "dry run does not write files",
			setup: func(t *dirTester, b *Builder) {
				b.WithFilesystem(afero.NewMemMapFs())
			},
			validate: func(t *dirTester, b *Builder) {
				t.assertDirEmpty()
			},
		},
		{
			name: "illegal directory traversals in rendered filenames are detected",
			setup: func(t *dirTester, b *Builder) {
				b.AddValues(template.Values{"filename": "../../"})
			},
			expectedErr: errors.New(`templated filename "{{.Values.filename}}" injected illegal directory traversal: ../../`),
		},
		{
			name: "rendering empty filename fails",
			setup: func(t *dirTester, b *Builder) {
				b.AddValues(template.Values{"filename": ""})
			},
			expectedErr: errors.New(`templated filename "{{.Values.filename}}" resolved to an empty string`),
		},
		{
			name: "does not overwrite existing files",
			setup: func(t *dirTester, b *Builder) {
				b.WithLicense(&license.Info{Body: `some license [fullname] [year]`})

				t.mustWriteFile("LICENSE", `do not touch`)
				t.mustWriteFile("README.md", `do not touch`)
			},
			validate: func(t *dirTester, b *Builder) {
				t.assertFileContains("LICENSE", `do not touch`)
				t.assertFileContains("README.md", `do not touch`)
			},
		},
		{
			name: "does not overwrite existing files selectively if OverwriteFiles is provided",
			setup: func(t *dirTester, b *Builder) {
				b.WithLicense(&license.Info{Body: `some license [fullname] [year]`}).
					OverwriteFiles([]string{
						"README.md",
						"./foobar/../foobar/somefile.yaml",
					})

				t.mustWriteFile(filepath.Join("foobar", "somefile.yaml"), `do not touch`)
				t.mustWriteFile("LICENSE", `do not touch`)
				t.mustWriteFile("README.md", `please overwrite`)
			},
			validate: func(t *dirTester, b *Builder) {
				t.assertFileContains(filepath.Join("foobar", "somefile.yaml"), "---\nsomekey: {{.Values.somekey}}\n")
				t.assertFileContains("LICENSE", `do not touch`)
				t.assertFileNotContains("README.md", `please overwrite`)
			},
		},
		{
			name: "files to overwrite must be relative",
			setup: func(t *dirTester, b *Builder) {
				b.OverwriteFiles([]string{
					"README.md",
					"./relfile",
					"/tmp/somefile",
				})
			},
			expectedErr: errors.New("found absolute path in overwrites: /tmp/somefile"),
		},
		{
			name: "does overwrite existing files if OverwriteAll is set",
			setup: func(t *dirTester, b *Builder) {
				b.WithLicense(&license.Info{Body: `some license [fullname] [year]`}).
					OverwriteAll(true)

				t.mustWriteFile("LICENSE", `please overwrite`)
				t.mustWriteFile("README.md", `please overwrite`)
			},
			validate: func(t *dirTester, b *Builder) {
				t.assertFileNotContains("LICENSE", `please overwrite`)
				t.assertFileNotContains("README.md", `please overwrite`)
			},
		},
		{
			name: "errors during configuration prevent build",
			setup: func(t *dirTester, b *Builder) {
				b.err = errors.New("whoops")
			},
			expectedErr: errors.New("whoops"),
		},
		{
			name: "template errors are returned",
			setup: func(t *dirTester, b *Builder) {
				b.AddValues(template.Values{"travis": "invalid"})
			},
			expectedErr: errors.New(`failed to render template: template: :4:13: executing "" at <.Values.travis.enabled>: can't evaluate field enabled in type interface {}`),
		},
		{
			name: "file copy errors are returned",
			setup: func(t *dirTester, b *Builder) {
				b.AddFile(badFile("badfile"))
			},
			expectedErr: errors.New("badfile"),
		},
		{
			name: "template read errors are returned",
			setup: func(t *dirTester, b *Builder) {
				b.AddFile(badFile("badtemplate.skel"))
			},
			expectedErr: errors.New("badtemplate.skel"),
		},
		{
			name: "errors while resolving templated filenames are returned",
			setup: func(t *dirTester, b *Builder) {
				b.AddValues(template.Values{"filename": func() {}})
			},
			expectedErr: errors.New(`failed to resolve templated filename "{{.Values.filename}}": failed to render template: template: :1:2: executing "" at <{{.Values.filename}}>: can't print {{.Values.filename}} of type func()`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()

			repo, err := repository.New("../testdata/repos/repo1")
			require.NoError(t, err)

			skeleton, err := repository.LoadSkeleton(ctx, repo, "advanced")
			require.NoError(t, err)

			tmpdir, err := ioutil.TempDir("", "kickoff-")
			require.NoError(t, err)

			defer os.RemoveAll(tmpdir)

			tester := &dirTester{T: t, dir: tmpdir}

			b := NewBuilder(test.config).
				AddValues(skeleton.Values)

			for _, f := range skeleton.Files {
				b.AddFile(f)
			}

			if test.setup != nil {
				test.setup(tester, b)
			}

			_, err = b.Build(tmpdir)
			if test.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, test.expectedErr, err)
			} else {
				require.NoError(t, err)
			}

			if test.validate != nil {
				test.validate(tester, b)
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
	contents, err := ioutil.ReadFile(t.path(file))
	require.NoError(t, err)
	assert.Equal(t, expectedContent, string(contents))
}

func (t *dirTester) assertFileNotContains(file, expectedContent string) {
	contents, err := ioutil.ReadFile(t.path(file))
	require.NoError(t, err)
	assert.NotEqual(t, expectedContent, string(contents))
}

func (t *dirTester) assertFileEmpty(file string) {
	contents, err := ioutil.ReadFile(t.path(file))
	require.NoError(t, err)
	assert.Len(t, contents, 0)
}

func (t *dirTester) assertFileExists(file string) {
	assert.FileExists(t, t.path(file))
}

func (t *dirTester) assertFileAbsent(file string) {
	_, err := ioutil.ReadFile(t.path(file))
	assert.True(t, os.IsNotExist(err))
}

func (t *dirTester) assertDirEmpty() {
	infos, err := ioutil.ReadDir(t.dir)
	require.NoError(t, err)
	assert.Len(t, infos, 0)
}

func (t *dirTester) mustWriteFile(file, content string) {
	path := t.path(file)

	err := os.MkdirAll(filepath.Dir(path), 0777)
	require.NoError(t, err)

	err = ioutil.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
}
