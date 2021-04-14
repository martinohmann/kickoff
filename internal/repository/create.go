package repository

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/martinohmann/kickoff/internal/file"
	"github.com/martinohmann/kickoff/internal/kickoff"
	log "github.com/sirupsen/logrus"
)

// Create creates a new repository at path and returns a reference to it.
func Create(path string) (*kickoff.RepoRef, error) {
	ref, err := kickoff.ParseRepoRef(path)
	if err != nil {
		return nil, err
	}

	if ref.IsRemote() {
		return nil, errors.New("creating remote repositories is not supported")
	}

	localPath := ref.LocalPath()

	if file.Exists(localPath) {
		return nil, fmt.Errorf("cannot create local repository: path %q already exists", localPath)
	}

	log.WithField("path", localPath).Info("creating skeleton repository")

	if err := os.MkdirAll(ref.SkeletonsPath(), 0755); err != nil {
		return nil, fmt.Errorf("failed to create repository in %q: %w", localPath, err)
	}

	return ref, nil
}

// CreateSkeleton creates a new skeleton with name in the referenced
// repository. Skeleton creation will fail with an error if ref does not
// reference a local repository. The created skeleton contains an example
// .kickoff.yaml and example README.md.skel as starter. Returns an error if
// creating path or writing any of the files fails.
func CreateSkeleton(ref *kickoff.RepoRef, name string) error {
	if name == "" {
		return errors.New("skeleton name must not be empty")
	}

	if ref.IsRemote() {
		return errors.New("creating skeletons in remote repositories is not supported")
	}

	localPath := ref.LocalPath()

	if !file.Exists(localPath) {
		return fmt.Errorf("cannot create skeleton: local repository %q does not exist", localPath)
	}

	path := ref.SkeletonPath(name)

	if file.Exists(path) {
		return fmt.Errorf("skeleton %q already exists in repository %q", name, ref.Name)
	}

	log.WithField("path", path).Info("creating directory")

	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create skeleton in %q: %w", path, err)
	}

	return writeFiles(path)
}

// CreateWithSkeleton creates a new local repository at path and seeds it with
// a skeleton.
func CreateWithSkeleton(path, skeletonName string) error {
	ref, err := Create(path)
	if err != nil {
		return err
	}

	return CreateSkeleton(ref, skeletonName)
}

func writeFiles(dir string) error {
	filenames := make([]string, 0, len(fileTemplates))
	for filename := range fileTemplates {
		filenames = append(filenames, filename)
	}

	sort.Strings(filenames)

	for _, filename := range filenames {
		path := filepath.Join(dir, filename)
		contents := fileTemplates[filename]

		log.WithField("path", path).Debug("creating skeleton file")

		err := ioutil.WriteFile(path, []byte(contents), 0644)
		if err != nil {
			return fmt.Errorf("failed to write skeleton file: %w", err)
		}
	}

	return nil
}

// fileTemplates is a mapping between filenames and the contents for these
// files when generating a new skeleton.
var fileTemplates = map[string]string{
	kickoff.SkeletonConfigFileName: `---
# Refer to the .kickoff.yaml documentation at https://kickoff.run/skeletons/configuration
# for a complete list of available skeleton configuration options.
#
# ---
# description: |
#   Some optional description of the skeleton that might be helpful to users.
# values:
#   myVar: 'myValue'
#   other:
#     someVar: false
`,
	"README.md.skel": `# {{.Project.Name}}

{{ if .License -}}
![GitHub](https://img.shields.io/github/license/{{.Project.Owner}}/{{.Project.Name}}?color=orange)

## License

The source code of {{.Project.Name}} is released under the {{.License.Name}}. See the bundled
LICENSE file for details.
{{- end }}
`,
}
