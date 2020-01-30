package skeleton

var (
	defaultReadmeSkeletonBytes = []byte(`# {{.Project.Name}}

{{ if .Values.travis.enabled -}}
[![Build Status](https://travis-ci.org/{{.Project.Owner}}/{{.Project.Name}}.svg?branch=master)](https://travis-ci.org/{{.Project.Owner}}/{{.Project.Name}})
{{- end }}
![GitHub](https://img.shields.io/github/license/{{.Project.Owner}}/{{.Project.Name}}?color=orange)

{{ if .License -}}
## License

The source code of {{.Project.Name}} is released under the {{.License.Name}}. See the bundled
LICENSE file for details.
{{- end }}
`)

	defaultConfigBytes = []byte(`---
# Description
# ===========
#
# Explain what this skeleton is about.
#
description: ""

# Parent
# ======
# 
# Skeletons can have parents which they inherit values and files from. Files
# and values of the same name in a child take precedence over those present in a
# parent.
#
# Example (parent in the same skeleton repo):
# -------------------------------------------
#
#   parent:
#     skeletonName: myparent
#
# Example (parent in local skeleton repo):
# ----------------------------------------
#
#   parent:
#     repositoryURL: /path/to/local/repo
#     skeletonName: myparent
#
# Example (parent in remote skeleton repo):
# -----------------------------------------
#
#   parent:
#     repositoryURL: https://github.com/martinohmann/kickoff-skeletons?rev=master
#     skeletonName: default
#
parent: null

# Custom configuration values
# ===========================
#
# Custom config is made available in *.skel template under {{ .Values }}. The
# values can be overridden on project creation.
#
# Example:
# --------
#
#   values:
#     travis:
#       enabled: false
#
values:
  travis:
    enabled: false
`)
)
