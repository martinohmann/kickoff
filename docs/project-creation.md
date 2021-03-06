---
title: Project creation
permalink: /project-creation
nav_order: 4
---

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
$ kickoff project create <name> <skeleton-name> [<skeleton-name>...] [flags]
```

It expects at least two arguments: the project name and the name of at least
one skeleton to create the project from. The new project is created in
`$PWD/<name>`.

**Note:** You can also pass a different directory via the `--dir` flag.

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

## Interactive mode

When `kickoff project create` is called with the `-i/--interactive` flag it
prompts for all project configuration options:

```bash
$ kickoff project create

? Select one or more project skeletons default:default
? Project name myproject
? Project directory [? for help, tab for suggestions] (/tmp/myproject)
[...]
```

## Overriding skeleton values

Skeletons can make use of custom values which can be overridden by the user
upon project creation. Available values (together with their defaults) can be
listed by inspecting the skeleton:

```bash
$ kickoff skeleton show myskeleton
```

Using the `--set` and `--values` flags you can override these:

```bash
$ kickoff project create myproject myskeleton --set someOtherKey.someNestedKey=43
$ kickoff project create myproject myskeleton --value values.yaml 
```

Refer to the [Accessing and setting template
variables](/skeletons/templating#accessing-and-setting-template-variables) and
[Configuring default values](/configuration#configuring-default-values)
documentation for more information.


## Including a `LICENSE`

Kickoff can automatically add a `LICENSE` file containing a popular open source
license which is obtained via the [GitHub Licenses
API](https://docs.github.com/en/rest/reference/licenses).

To add a license to your project, just specify its name using the `--license` flag:

```bash
$ kickoff project create myproject myskeleton --license MIT
```

It will automatically fill in the year and project owner into the license if
these fields are supported by the license, e.g.:

```bash
$ cat ~/myproject/LICENSE

MIT License

Copyright (c) 2020 johndoe

Permission is hereby granted, free of charge, to any person obtaining a copy
[...]
```

For a list of available licenses run:

```bash
$ kickoff licenses list
```

You can also [configure a default project
license](/configuration#configuring-a-default-project-license) that will be
used for all new projects if not explicitly overridden.

## Including a `.gitignore`

You can automatically include a `.gitignore` file with your project which can
be built from one or multiple gitignore templates available via the [GitHub
Gitignores API](https://docs.github.com/en/rest/reference/gitignore). The
templates can be passed as comma separated list via the `--gitignore` flag:

```bash
$ kickoff project create myproject myskeleton --gitignore go,hugo
```

For a list of available `.gitignore` templates run:

```bash
$ kickoff gitignores list
```

It is also possible to [configure default `.gitignore`
templates](/configuration#configuring-default-project-gitignore-templates)
which can be overridden explicitly on project creation.

## Creating a project from multiple skeletons

Projects can be created by composing multiple skeletons together. This is just
as simple as providing multiple skeletons instead of one after the project name
on project creation:

```bash
$ kickoff project create myproject repo:skeleton1 otherrepo:skeleton2 skeleton3
```

Note that the skeletons are merged left to right, so files and values from
skeletons on the right will override files and values of the same name from
other skeletons to the left.

**Caveats:** Creating project from skeletons that are defining the same value
with different types might cause unexpected behaviour or even fail while
rendering templates.

