---
title: Project creation
nav_order: 4
---

{% include wip.md %}

# Project creation
{: .no_toc}

In the [Getting started guide](getting-started) you already encountered a basic
example of how to create a project from a project skeleton. This section
explains the features of the project creation command in more detail. Amongst
other things you will learn how to override [skeleton template
values](/skeletons/templating) and how to automatically add a .gitignore or
LICENSE file to your new project.

1. TOC
{:toc}

## Basics

The basic usage of the project creation command looks like this:

```bash
$ kickoff project create [repo:]skeleton project-dir [flags...]
```

It expects two arguments: the skeleton to create the project from and the
directory where the project should be created.

The skeleton needs to be prefixed with the name of the repository if the
skeleton name is ambiguous. For example:

- Given two repositories `repo1` and `repo2`, both containing a skeleton named
  `myskeleton`: if you want to create a new project using `myskeleton` from
  `repo2`, you need to pass `repo2:myskeleton` to the project creation command,
  otherwise kickoff cannot figure out which of the two `myskeleton` skeleton
  you want.
- If the skeleton is only present in one of the repositories that you have
  configured, the repository prefix can be omitted, for example specifying
  `myskeleton` would be enough for kickoff to know which skeleton to pick.

If you pass an ambiguous skeleton name, kickoff will let you know and print out
the options you have for referencing a skeleton explictly.

To learn more about available flags, have a look a the usage examples that are
displayed as part of the command help:

```bash
$ kickoff project create --help
```

## Overriding skeleton values

TODO

## Including a license

TODO

## Including a `.gitignore`

TODO

## Creating a project from multiple skeletons

TODO

