package project

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/martinohmann/kickoff/internal/config"
	"github.com/martinohmann/kickoff/internal/file"
	"github.com/martinohmann/kickoff/internal/license"
	"github.com/martinohmann/kickoff/internal/skeleton"
	"github.com/martinohmann/kickoff/internal/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	tests := []struct {
		name          string
		expectedErr   error
		createOptions *CreateOptions
		setup         func(t *testing.T, outputDir string)
		validate      func(t *testing.T, outputDir string)
	}{
		{
			name: "create with license and gitignore file",
			createOptions: &CreateOptions{
				Gitignore: "somegitignorebody",
				Config: config.Project{
					Owner: "johndoe",
				},
				License: &license.Info{
					Body: `some license [fullname] [year]`,
				},
				InitGit: true,
			},
			validate: func(t *testing.T, outputDir string) {
				if file.Exists(filepath.Join(outputDir, ".kickoff.yaml")) {
					assert.Fail(t, "expected file %s to be absent but it is not", filepath.Join(outputDir, ".kickoff.yaml"))
				}

				assert.FileExists(t, filepath.Join(outputDir, "README.md"))
				assert.FileExists(t, filepath.Join(outputDir, "foobar", "somefile.yaml"))
				assert.DirExists(t, filepath.Join(outputDir, ".git"))

				contentTests := []struct {
					path     string
					contents []byte
				}{
					{
						path:     filepath.Join(outputDir, ".gitignore"),
						contents: []byte(`somegitignorebody`),
					},
					{
						path:     filepath.Join(outputDir, "LICENSE"),
						contents: []byte(`some license johndoe ` + strconv.Itoa(time.Now().Year())),
					},
				}

				for _, test := range contentTests {
					buf, err := ioutil.ReadFile(test.path)
					require.NoError(t, err)

					if !bytes.Equal(buf, test.contents) {
						t.Fatalf(`expected %s to contain %q, but got %q`, test.path, test.contents, buf)
					}
				}
			},
		},
		{
			name: "dry run does not write files",
			createOptions: &CreateOptions{
				DryRun: true,
			},
			validate: func(t *testing.T, outputDir string) {
				infos, err := ioutil.ReadDir(outputDir)
				require.NoError(t, err)
				assert.Len(t, infos, 0)
			},
		},
		{
			name: "illegal directory traversals in rendered filenames are detected",
			createOptions: &CreateOptions{
				Values: template.Values{
					"filename": "../../",
				},
			},
			expectedErr: errors.New(`templated filename "{{.Values.filename}}" injected illegal directory traversal: ../../`),
		},
		{
			name: "rendering empty filename fails",
			createOptions: &CreateOptions{
				Values: template.Values{
					"filename": "",
				},
			},
			expectedErr: errors.New(`templated filename "{{.Values.filename}}" resolved to an empty string`),
		},
		{
			name: "does not overwrite existing files if Overwrite is false",
			createOptions: &CreateOptions{
				Overwrite: false,
				License: &license.Info{
					Body: `some license [fullname] [year]`,
				},
			},
			setup: func(t *testing.T, outputDir string) {
				require.NoError(t, ioutil.WriteFile(filepath.Join(outputDir, "README.md"), []byte(`do not touch`), 0644))
				require.NoError(t, ioutil.WriteFile(filepath.Join(outputDir, "LICENSE"), []byte(`do not touch`), 0644))
			},
			validate: func(t *testing.T, outputDir string) {
				contents, err := ioutil.ReadFile(filepath.Join(outputDir, "README.md"))
				require.NoError(t, err)
				assert.Equal(t, `do not touch`, string(contents))
				contents, err = ioutil.ReadFile(filepath.Join(outputDir, "LICENSE"))
				require.NoError(t, err)
				assert.Equal(t, `do not touch`, string(contents))
			},
		},
		{
			name: "does overwrite existing files if Overwrite is true",
			createOptions: &CreateOptions{
				Overwrite: true,
				License: &license.Info{
					Body: `some license [fullname] [year]`,
				},
			},
			setup: func(t *testing.T, outputDir string) {
				require.NoError(t, ioutil.WriteFile(filepath.Join(outputDir, "README.md"), []byte(`do not touch`), 0644))
				require.NoError(t, ioutil.WriteFile(filepath.Join(outputDir, "LICENSE"), []byte(`do not touch`), 0644))
			},
			validate: func(t *testing.T, outputDir string) {
				contents, err := ioutil.ReadFile(filepath.Join(outputDir, "README.md"))
				require.NoError(t, err)
				assert.NotEqual(t, `do not touch`, string(contents))
				contents, err = ioutil.ReadFile(filepath.Join(outputDir, "LICENSE"))
				require.NoError(t, err)
				assert.NotEqual(t, `do not touch`, string(contents))
			},
		},
		{
			name:          "does not create file if template rendered to an empty string",
			createOptions: &CreateOptions{},
			validate: func(t *testing.T, outputDir string) {
				_, err := ioutil.ReadFile(filepath.Join(outputDir, "optional-file"))
				require.True(t, os.IsNotExist(err))
			},
		},
		{
			name: "does create file if template rendered to an empty string and AllowEmpty is true",
			createOptions: &CreateOptions{
				AllowEmpty: true,
			},
			validate: func(t *testing.T, outputDir string) {
				contents, err := ioutil.ReadFile(filepath.Join(outputDir, "optional-file"))
				require.NoError(t, err)
				assert.Len(t, contents, 0)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			loader, err := skeleton.NewSingleRepositoryLoader("../testdata/repos/repo1")
			if err != nil {
				t.Fatal(err)
			}

			skeleton, err := loader.LoadSkeleton("advanced")
			if err != nil {
				t.Fatal(err)
			}

			tmpdir, err := ioutil.TempDir("", "kickoff-")
			if err != nil {
				t.Fatal(err)
			}

			defer os.RemoveAll(tmpdir)

			if test.setup != nil {
				test.setup(t, tmpdir)
			}

			err = Create(skeleton, tmpdir, test.createOptions)
			if test.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, test.expectedErr, err)
			} else {
				require.NoError(t, err)
			}

			if test.validate != nil {
				test.validate(t, tmpdir)
			}
		})
	}
}
