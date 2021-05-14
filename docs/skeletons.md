---
title: Skeletons
permalink: /skeletons
nav_order: 6
has_children: true
---

# Project skeletons

Project skeletons are blueprints that can be used to create new projects. They
can be fully templated and composed together.

This section provides documentation for creating, customizing and using project
skeletons.

The most basic commands for interacting with them are listed below without much
explanation. Navigate to the subpages for more details.

### Skeleton structure

In essence project skeletons are just simple directories. [Read
more](/skeletons/structure) about skeleton directory structure.

### Skeleton creation

Project skeletons can be created with kickoff very easily:

```bash
$ kickoff skeleton create myrepo myskeleton
```

However, there is more to it. Check out the [Creating project
skeletons](/skeletons/creating-skeletons) section to find out.

### Listing skeletons

This is how you list all skeletons that are available in the skeleton
repositories you configured:

```bash
$ kickoff skeleton list

Repository  Name
default     default
default     golang/cli
default     golang/library
```

To learn more about the working with repositories, head over to the
[repositories documentation](/repositories).

### Inspecting skeletons

In addition to listing skeletons, they can also be inspected individual:

```bash
$ kickoff skeleton show default:golang/cli

Repository  default
Name        golang/cli
Path        ~/.cache/kickoff/repositories/4c76fb4fd87cd5b1dca9d94fa35751b06f507109b75bd3a4bc35012ed33cecfb/skeletons/golang/cli

Files                       Values
golang/cli                  golang:
├── .github/                  targetVersion: "1.15"
│   └── workflows/
│       └── build.yml
├── .gitignore.skel
├── .golangci.yml
├── .goreleaser.yml.skel
├── Makefile.skel
├── README.md.skel
├── cmd/
│   └── {{.Project.Name}}/
│       └── main.go.skel
├── doc.go.skel
├── go.mod.skel
└── pkg/
    └── cmd/
        └── root.go.skel
```

To inspect the content of individual skeleton files, just add the file path to the `show` command:

```bash
$ kickoff skeleton show default:golang/cli doc.go.skel

// Package {{.Project.Name|goPackageName}} provides a library to work with
// stuff and things.
package {{.Project.Name|goPackageName}}
```
