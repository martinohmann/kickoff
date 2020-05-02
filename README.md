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
- [Using skeleton repositories](#using-skeleton-repositories)
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
curl -SsL -o kickoff "https://github.com/martinohmann/kickoff/releases/latest/download/kickoff_$(uname -s)_$(uname -m)"
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

The kickoff cli comes with detailed examples on each subcommand, so be sure to
check them out. Showing the help is a good starting point:

```bash
# List of all available commands.
kickoff help

# Help for the `project create` subcommand.
kickoff project create help
```

## Using skeleton repositories

Kickoff supports local and remote skeleton repositories. If you want, you can
use the repository that come along with `kickoff`. Head over to the
[kickoff-skeletons](https://github.com/martinohmann/kickoff-skeletons)
repository for ready-to-use skeletons and to get some inspiration to create
your own.

You can add the `kickoff-skeletons` repository to your config to directly
create projects from the available skeletons:

```bash
kickoff repository add kickoff-skeletons https://github.com/martinohmann/kickoff-skeletons
```

### Local skeleton repositories

Kickoff supports local repositories which do not necessarily need to be git
repos. If you did not create a local repository via `kickoff init`, you can
create one like this:

```bash
kickoff repository create ~/path/to/new/repo
kickoff repository add myrepo ~/path/to/new/repo
```

The `kickoff repository create` command will create a new repository which
already contains a minimal `default` skeleton with a commented `.kickoff.yaml`
file and a `README.md.skel` skeleton to get you started. You can delete it or
customize it to your needs.

You can verify that your local repository was correctly created and added by
listing the available kickoff repositories:

```bash
kickoff repository list
```

### Remote skeleton repositories

Add a remote skeleton repository and create a new project:

```bash
kickoff repository add myremoterepo https://github.com/myuser/myskeletonrepo?revision=v1.0.0
kickoff repository list
kickoff project create myremoterepo:myskeleton ~/path/to/my/new/project
```

Remote repository urls can contain an optional `revision` query parameter which
may point to a commit, tag or branch. If omitted `master` is assumed.

## Project Skeletons

A skeleton is just a subdirectory of the `skeletons/` directory inside your
local repository. The skeleton directory must contain a `.kickoff.yaml` file
(which may be empty).

### Creating skeletons in a local repository

Creating a new skeleton with some boilerplate can be done like this:

```bash
kickoff skeleton create ~/path/to/local/skeleton-repository/skeletons/myskeleton
```

You can verify it by listing all available skeletons:

```bash
kickoff skeleton list
```

### Skeleton templating

Any file with the `.skel` extension is treated as a [Go
templates](https://golang.org/pkg/text/template/) and can be fully templated.
When creating a project the resolved `.skel` templates will be written to the
target directory with the `.skel` extension stripped, e.g. `README.md.skel`
becomes `README.md`.

### Template variables

Kickoff makes available a couple of variables to these templates:

| Variable                 | Description                                                                                                                                                                                                                               |
| ---                      | ---                                                                                                                                                                                                                                       |
| `.Project.Host`          | The git host you specified during `kickoff init`, e.g. `github.com`                                                                                                                                                                       |
| `.Project.Owner`         | The project owner you specified, e.g. `martinohmann`                                                                                                                                                                                      |
| `.Project.Email`         | The project email you specified, e.g. `foo@bar.baz`                                                                                                                                                                                       |
| `.Project.Name`          | The name you specified when running `kickoff project create`                                                                                                                                                                              |
| `.Project.License`       | The name of the license, if you picked one                                                                                                                                                                                                |
| `.Project.Gitignore`     | Comma-separated list of gitignore templates, if provided                                                                                                                                                                                  |
| `.Project.URL`           | The URL to the project repo, e.g. `https://github.com/martinohmann/myproject`                                                                                                                                                             |
| `.Project.GoPackagePath` | The package path for go projects, e.g. `github.com/martinohmann/myproject`                                                                                                                                                                |
| `.Project.Author`        | If an email is present, this will resolve to `project-owner <foo@bar.baz>`, otherwise just `owner`                                                                                                                                        |
| `.Values`                | The merged result of the `values` from your local kickoff `config.yaml`, the project skeleton's `values` from `.kickoff.yaml`, values from files passed via `--values` and any values that were set using `--set` during project creation |

### Templating file and directory names

Kickoff will try to resolve Go template variables in file and directory names.
E.g. a directory `cmd/{{.Project.Name}}` will be resolved to `cmd/myproject` if
the project is named `myproject`. It is also possible to put arbitrary files
into these directories. These will be moved to the correct place after the
directory name was resolved. You can check out this example skeleton which
makes use of this feature:
[kickoff-skeletons:golang/cli](https://github.com/martinohmann/kickoff-skeletons/tree/master/skeletons/golang/cli).

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
