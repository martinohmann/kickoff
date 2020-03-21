# kickoff

[![Build Status](https://travis-ci.com/martinohmann/kickoff.svg?branch=master)](https://travis-ci.com/martinohmann/kickoff)
[![codecov](https://codecov.io/gh/martinohmann/kickoff/branch/master/graph/badge.svg)](https://codecov.io/gh/martinohmann/kickoff)
[![GoDoc](https://godoc.org/github.com/martinohmann/kickoff?status.svg)](https://godoc.org/github.com/martinohmann/kickoff)
![GitHub](https://img.shields.io/github/license/martinohmann/kickoff?color=orange)

Bootstrap projects from skeleton directories.

[![asciicast](https://asciinema.org/a/T53cAY9Uitt4I8XQT5rWPKDxk.svg)](https://asciinema.org/a/T53cAY9Uitt4I8XQT5rWPKDxk)

Contents
--------

- [Features](#features)
- [Installation](#installation)
- [Quickstart](#quickstart)
- [Using remote skeleton repositories](#using-remote-skeleton-repositories)
- [Project Skeletons](#project-skeletons)
- [Environment variables](#environment-variables)
- [Shell completion](#shell-completion)
- [Skeleton inheritance](#skeleton-inheritance)
- [Skeleton composition](#skeleton-composition)

## Features

- Templating of filenames, directory names and file contents via [Go
  templates](https://golang.org/pkg/text/template/) and
  [Sprig](http://masterminds.github.io/sprig/).
- Extensible by allowing users to pass arbitrary values to templates via config
  files or CLI flags.
- Automatically populate LICENSE file with an open source license obtained from
  the [GitHub Licenses API](https://developer.github.com/v3/licenses/).
- Automatically add a .gitignore created from templates obtained from
  [gitignore.io](https://gitignore.io).
- Set local author, repository and skeleton defaults using custom config file.
- Dry run for project creation.
- Skeleton inheritance: skeletons can inherit files and values from an optional
  parent skeleton.
- Skeleton composition: projects can be created by composing multiple skeletons
  together. This is similar to inheritance but allows for way more flexible use
  of skeletons.

## Installation

### From binary release

Currently only Linux and MacOSX are packaged as binary releases. Check out the
[releases](https://github.com/martinohmann/kickoff/releases) for all available
versions.

```bash
curl -SsL -o kickoff "https://github.com/martinohmann/kickoff/releases/download/v0.0.1/kickoff_0.0.1_$(uname -s | tr '[:upper:]' '[:lower:]')_x86_64"
chmod +x kickoff
mv kickoff $GOPATH/bin/
```

### From source

```bash
git clone https://github.com/martinohmann/kickoff
cd kickoff
make install
```

This will install the `kickoff` binary to `$GOPATH/bin/kickoff`.

You can verify the installation by printing the version:

```bash
kickoff version
```

## Quickstart

Initialize the kickoff config and create a new project:

```bash
kickoff init
kickoff project create default ~/path/to/my/new/project --license mit --gitignore go,hugo
```

## Using remote skeleton repositories

Add a remote skeleton repository and create a new project:

```bash
kickoff repository add myremoterepo https://github.com/myuser/myskeletonrepo?revision=v1.0.0
kickoff repository list
kickoff project create myremoterepo:myskeleton ~/path/to/my/new/project
```

Remote repository urls can contain an optional `revision` query parameter which
may point to a commit, tag or branch. If omitted `master` is assumed.

## Project Skeletons

Head over to the
[kickoff-skeletons](https://github.com/martinohmann/kickoff-skeletons)
repository for ready-to-use skeletons and to get some inspiration to create
your own.

You can add the `kickoff-skeletons` repository to your config to directly
create projects from the available skeletons:

```bash
kickoff repository add kickoff-skeletons https://github.com/martinohmann/kickoff-skeletons
```

## Environment variables

The following environment variables can be used to configure kickoff:

| Name             | Description                                                                                          |
| ---              | ---                                                                                                  |
| `KICKOFF_CONFIG` | Path to the kickoff config. Can be overridden with the `--config` flag.                              |
| `KICKOFF_EDITOR` | Editor used by `kickoff config edit`. If unset, `EDITOR` environment will be used. Fallback is `vi`. |

## Shell completion

Add to your `~/.bashrc` for bash completion:

```bash
. <(kickoff completion bash)
```

Add to your `~/.zshrc` for zsh completion:

```bash
. <(kickoff completion zsh)
```

## Skeleton inheritance

Skeletons can inherit from other skeletons. Just add the `parent` configuration
to the `.kickoff.yaml` of the skeleton like this:

```yaml
parent:
  repositoryURL: https://github.com/martinohmann/kickoff-skeletons?revision=master
  skeletonName: my-parent-skeleton
```

If `repositoryURL` is omitted, the same repository as the one of the skeleton
is assumed. `repositoryURL` can be a remote URL or local path. Remote
repository urls can contain an optional `revision` query parameter which may
point to a commit, tag or branch. If omitted `master` is assumed.

## Skeleton composition

Projects can be created by composing multiple skeletons together. This is just
as simple as providing multiple skeletons instead of one as comma separated
list on project creation:

```bash
kickoff project create skeleton1,skeleton2,skeleton3 /path/to/project
```

Note that the skeletons are merged left to right, so files and values from
skeletons on the right will override files and values of the same name from
other skeletons.

## License

The source code of kickoff is released under the MIT License. See the bundled
LICENSE file for details.
