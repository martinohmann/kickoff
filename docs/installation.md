---
title: Installation
nav_order: 2
---

# Installation
{: .no_toc }

Kickoff can be installed in multiple ways, choose the one that suits you best.

1. TOC
{:toc}


## Installing Kickoff from the latest release

This is the recommended way to install Kickoff. Currently only Linux and MacOSX
are packaged as binary releases. Check out the
[releases](https://github.com/martinohmann/kickoff/releases) for all available
versions.

Download the latest release binary and make it executable:

```bash
$ curl -SsL -o kickoff "https://github.com/martinohmann/kickoff/releases/latest/download/kickoff_$(uname -s)_$(uname -m)"
$ chmod +x kickoff
```

Move it somewhere into your `$PATH`, e.g. `$GOPATH/bin` if you have set that
up, or install it globalling into `/usr/local/bin` (possibly required `sudo`):

```bash
$ mv kickoff $GOPATH/bin/
```

If everything is correctly setup you should be able to print the Kickoff version:

```bash
$ kickoff version
```

## Installing Kickoff from source

You can also build the `kickoff` binary from source. This assumes that you have
`make` and `go` installed in a suitable version and that your `$GOPATH` is
setup properly.

First clone the repository and change into it:

```bash
$ git clone https://github.com/martinohmann/kickoff
$ cd kickoff
```

**Optional**: if you want to build a specific version, you should check it out first:

```bash
$ git checkout <version>
```

Refer to the [kickoff releases
page](https://github.com/martinohmann/kickoff/releases) for available tags.

Build the `kickoff` binary:

```bash
$ make build
```

Move it somewhere into your path, or run `make install` to install it to
`$GOPATH/bin/kickoff`.

You can verify that the installation was successful by printing the kickoff
version:

```bash
$ kickoff version
```

## Help commands

The kickoff CLI comes with detailed examples on each subcommand, so be sure to
check them out. Showing the help is a good starting point:


```bash
# List of all available commands.
$ kickoff help

# Help for the `project create` subcommand.
$ kickoff project create help
```

## Next steps

* [Getting started](getting-started): Initialize the Kickoff configuration and
  create your first project from a project skeleton.
