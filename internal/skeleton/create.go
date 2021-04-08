package skeleton

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/martinohmann/kickoff/internal/kickoff"
	log "github.com/sirupsen/logrus"
)

// Create creates a new skeleton at path. The created skeleton contains an
// example .kickoff.yaml and example README.md.skel as starter. Returns an
// error if creating path or writing any of the files fails.
func Create(path string) error {
	log.WithField("path", path).Info("creating directory")

	err := os.MkdirAll(path, 0755)
	if err != nil {
		return fmt.Errorf("failed to create skeleton dir %q", err)
	}

	return writeFiles(path)
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

		log.WithField("path", path).Info("creating file")

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
