# kickoff

[![Build Status](https://travis-ci.com/martinohmann/kickoff.svg?branch=master)](https://travis-ci.com/martinohmann/kickoff)
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

## Creating a project from a skeleton

Quick example:

```
kickoff project create myskeleton ~/path/to/my/new/project --license mit
```

## Project Skeletons

Head over to the
[kickoff-skeletons](https://github.com/martinohmann/kickoff-skeletons)
repository for ready-to-use skeletons and to get some inspiration to create
your own.

## License

The source code of kickoff is released under the MIT License. See the bundled
LICENSE file for details.
