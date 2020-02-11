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

	"github.com/martinohmann/kickoff/pkg/config"
	"github.com/martinohmann/kickoff/pkg/file"
	"github.com/martinohmann/kickoff/pkg/license"
	"github.com/martinohmann/kickoff/pkg/skeleton"
	"github.com/martinohmann/kickoff/pkg/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	tests := []struct {
		name          string
		expectedErr   error
		createOptions *CreateOptions
		validate      func(t *testing.T, outputDir string)
	}{
		{
			name: "create with license and gitignore file",
			createOptions: &CreateOptions{
				Gitignore: "somegitignorebody",
				Config: config.Project{
					Owner: "johndoe",
					Email: "john@example.com",
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
						contents: []byte(`some license johndoe <john@example.com> ` + strconv.Itoa(time.Now().Year())),
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
