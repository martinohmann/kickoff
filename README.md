# kickoff -- bootstrap projects from skeleton directories

[![Build Status](https://travis-ci.com/martinohmann/kickoff.svg?branch=master)](https://travis-ci.com/martinohmann/kickoff)
[![codecov](https://codecov.io/gh/martinohmann/kickoff/branch/master/graph/badge.svg)](https://codecov.io/gh/martinohmann/kickoff)
[![Go Report Card](https://goreportcard.com/badge/github.com/martinohmann/kickoff)](https://goreportcard.com/report/github.com/martinohmann/kickoff)
[![GoDoc](https://godoc.org/github.com/martinohmann/kickoff?status.svg)](https://godoc.org/github.com/martinohmann/kickoff)
![GitHub](https://img.shields.io/github/license/martinohmann/kickoff?color=orange)

Kickoff makes starting new projects easy and fast. Use community project
skeletons or create your own. No more need to copy & paste initial boilerplate
like Makefiles, CI configuration or language specific configuration files from
existing projects to a new one.

[![asciicast](https://asciinema.org/a/T53cAY9Uitt4I8XQT5rWPKDxk.svg)](https://asciinema.org/a/T53cAY9Uitt4I8XQT5rWPKDxk)

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

## Documentation

Head over to the [kickoff documentation](https://kickoff.run) or directly jump
into the [Getting started guide](https://kickoff.run/getting-started).

## License

The source code of kickoff is released under the MIT License. See the bundled
LICENSE file for details.
