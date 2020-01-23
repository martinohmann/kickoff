// Package boilerplate contains the string representations of the default
// kickoff config, the default skeleton config and an example README template
// for initializing a new skeleton.
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
#     email: john@example.com
#
project:
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

  # Gitignore configuration
  # =======================
  #
  # A .gitignore file will be automatically generated if gitignore templates are
  # specified. All gitignore templates available via the gitignore.io API are
  # supported.
  #
  # Check out the 'kickoff gitignore list' for a list a available templates.
  #
  # Example:
  # --------
  #
  #   gitignore: go
  #
  # Multiple gitignore templates:
  # -----------------------------
  #
  #   gitignore: go,helm,hugo
  #
  gitignore: ""

  # Host/owner configuration
  # ========================
  #
  # The host/owner configuration is made available to *.skel templates so you
  # can build links related to your project, e.g. for CI badges or
  # documentation links.
  #
  # If empty, kickoff will attempt to fetch the SCM owner from the github.user
  # or user.name git # config (if present).
  #
  # Example:
  # --------
  #
  #   host: github.com (this is the default if the field is empty)
  #   owner: johndoe
  #
  owner: ""

# Skeleton repository configuration
# =================================
#
# The repositories config key is a map of repository name to location of the
# skeleton repository. This can be a local dir or a local/remote git
# repository. If no repository with the name "default" is specified, it will be
# automatically added and pointed to <local-config-dir>/kickoff/repository, if
# that directory exists.
#
# Local dir example:
# ------------------
#
#   repositories:
#     my-local-dir: /path/to/my/skeletons/dir
#
# Local repo (with branch) example:
# ---------------------------------
#
#   repositories:
#     my-local-repo: /path/to/my/skeletons/repo?branch=develop
#
# Remote repo (with branch) example:
# ----------------------------------
#
#   repositories:
#     my-remote-repo: https://github.com/myuser/myskeletonrepo?branch=develop
#
# Multiple repos example:
# -----------------------
#
#   repositories:
#     my-local-dir: /path/to/my/skeletons/dir
#     my-local-repo: /path/to/my/skeletons/repo?branch=develop
#     my-remote-repo: https://github.com/myuser/myskeletonrepo?branch=develop
#
repositories: {}

# Parent skeleton
# ===============
#
# If parent is not null, the references skeleton's files and values will be
# merged into the skeleton.
#
# Example:
# --------
#
#   parent:
#     repositoryURL: https://github.com/martinohmann/kickoff-skeletons
#     skeletonName: my-parent-skeleton
#
parent: null

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
# Description
# ===========
#
# Explain what this skeleton is about
#
description: ""

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
