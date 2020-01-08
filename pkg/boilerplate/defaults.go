package boilerplate

const (
	DefaultReadmeText = `{{.Project.Name}}
=====

{{ if .Values.travis.enabled -}}
[![Build Status](https://travis-ci.org/{{.Git.User}}/{{.Git.RepoName}}.svg?branch=master)](https://travis-ci.org/{{.Git.User}}/{{.Git.RepoName}})
{{- end }}
![GitHub](https://img.shields.io/github/license/{{.Git.User}}/{{.Git.RepoName}}?color=orange)

{{ if .License -}}
License
-------

The source code of {{.Project.Name}} is released under the {{.License.Name}}. See the bundled
LICENSE file for details.
{{- end }}
`

	DefaultConfigText = `---
## Custom configuration values
#
# Custom config is made available in *.skel template under {{ .Values }}. The
# values can be overridden on project creation.
values:
  travis:
    enabled: false
`
)

func DefaultReadme() []byte {
	return []byte(DefaultReadmeText)
}

func DefaultConfig() []byte {
	return []byte(DefaultConfigText)
}
