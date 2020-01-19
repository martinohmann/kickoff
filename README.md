# kickoff

[![Build Status](https://travis-ci.com/martinohmann/kickoff.svg?branch=master)](https://travis-ci.com/martinohmann/kickoff)
[![codecov](https://codecov.io/gh/martinohmann/kickoff/branch/master/graph/badge.svg)](https://codecov.io/gh/martinohmann/kickoff)
[![GoDoc](https://godoc.org/github.com/martinohmann/kickoff?status.svg)](https://godoc.org/github.com/martinohmann/kickoff)
![GitHub](https://img.shields.io/github/license/martinohmann/kickoff?color=orange)

Bootstrap projects from skeletons. Documentation is currently WIP.

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

## Installation

Quick:

```
go get -u github.com/martinohmann/kickoff/cmd/kickoff
```

Recommended:

```
git clone https://github.com/martinohmann/kickoff
make install
```

Verify installation by running:

```
kickoff version
```

## Quickstart

Initialize the kickoff config and create a new project:

```
kickoff init
kickoff project create default ~/path/to/my/new/project --license mit --gitignore go,hugo
```

## Using remote skeleton repositories

Add a remote skeleton repository and create a new project:

```
kickoff repository add myremoterepo https://github.com/myuser/myskeletonrepo
kickoff repository list
kickoff project create myremoterepo:myskeleton ~/path/to/my/new/project
```

## Project Skeletons

Head over to the
[kickoff-skeletons](https://github.com/martinohmann/kickoff-skeletons)
repository for ready-to-use skeletons and to get some inspiration to create
your own.

You can add the `kickoff-skeletons` repository to your config to directly
create projects from the available skeletons:

```
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

```
. <(kickoff completion bash)
```

Add to your `~/.zshrc` for zsh completion:

```
. <(kickoff completion zsh)
```

## License

The source code of kickoff is released under the MIT License. See the bundled
LICENSE file for details.
