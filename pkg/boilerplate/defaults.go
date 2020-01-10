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
# The config file provides defaults which can be explicitly overridden via CLI
# flags.

# Project configuration
# =====================
#
# Configures the author and email address that are automatically injected in
# license texts and are also available in *.skel templates.
#
# If empty, kickoff will try to populate project.author and project.email from
# the user.fullname and user.email git config attributes (if they exist).
#
# Example:
# --------
#
#   project:
#     author: John Doe
#     email: john@example.com
#
project:
  author: ""
  email: ""

# License configuration
# =====================
#
# A LICENSE file will be automatically generated if a license is specified. All
# licenses available via the GitHub Licenses API are supported.
#
# Check out the 'kickoff licenses list' for a list a available
# licenses.
#
# Example:
# --------
#
#   license: mit
#
license: ""

# Git configuration
# =================
#
# The git configuration is made available to *.skel templates so you can build
# links related to your project, e.g. for CI badges or documentation links.
#
# If empty, kickoff will attempt to fetch the git user from the github.user git
# config (if present).
#
# Example:
# --------
#
#   git:
#     host: github.com (this is the default if the field is empty)
#     user: johndoe
#
git:
  user: ""

# Skeleton configuration
# ======================
#
# The skeletons.repositoryURL specifies the location of the skeleton
# repository. This can be a local dir or a local/remote git repository.
#
# Local dir example:
# ------------------
#
#   skeletons:
#     repositoryURL: /path/to/my/skeletons/dir
#
# Local repo (with branch) example:
# ---------------------------------
#
#   skeletons
#     repositoryURL: /path/to/my/skeletons/repo?branch=develop
#
# Remote repo (with branch) example:
# ----------------------------------
#
#   skeletons:
#     repositoryURL: https://github.com/myuser/myskeletonrepo?branch=develop
#
skeletons:
  repositoryURL: ""

# Custom configuration values
# ===========================
#
# Custom config is made available in *.skel template under {{ .Values }}. E.g.
# the example below can be used in templates as
# {{ .Values.travis.publishCodeCov }}.
#
# Example:
# --------
#
#   values:
#     travis:
#       publishCodeCov: true
#
values: {}
`

	DefaultSkeletonConfigText = `---
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
values: {}
`
)

func DefaultReadmeBytes() []byte {
	return []byte(DefaultReadmeText)
}

func DefaultConfigBytes() []byte {
	return []byte(DefaultConfigText)
}

func DefaultSkeletonConfigBytes() []byte {
	return []byte(DefaultSkeletonConfigText)
}