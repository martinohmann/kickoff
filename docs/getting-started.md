---
title: Getting started
nav_order: 3
---

# Getting started
{: .no_toc }

This page walks you through the initial setup of Kickoff and will make you
familiar with the basic steps to create a project from a project skeleton.

If you did not install Kickoff yet, head over to the [installation
guide](installation) and do that first before you proceed.

1. TOC
{:toc}

## Initializing the Kickoff configuration

Kickoff needs some initial configuration to be useful. For creating projects
from project skeletons, it needs to know some basic information about the
default Git hosting platform (e.g. `github.com`) and your username there (e.g.
`johndoe`).

This information is necessary because it is made available to
skeleton templates. e.g. for building useful links to the repository, the CI
integration or to be used in language specific configuration that require a
project namespace.

You also need to configure a default project skeleton repository so that you
have something to work with.

Now let's get started by creating your initial Kickoff configuration:

```bash
$ kickoff init
```

This will launch an interactive shell that let's you set the initial
configuration values. Some options are already prefilled with values that
Kickoff detected from your git config or from an existing configuration file if
you already ran `kickoff init` in the past.

```bash
$ kickoff init

? Project host github.com
? Project owner johndoe
? Project email johndoe@example.com
? Do you want to set a default project license? No
? Do you want to select default .gitignore templates? No
```

You can type `?` and hit `Enter` for every configuration option to get some
additional information about it. The step where you get asked about the default skeleton repository is important:

```bash
? Default skeleton repository [? for help] (https://github.com/martinohmann/kickoff-skeletons)
```

This lets you specify the local or remote location of your default skeleton
repository. By default, this will use the remote repository at
[https://github.com/martinohmann/kickoff-skeletons](https://github.com/martinohmann/kickoff-skeletons)
that comes together with Kickoff.

Let's choose a local path instead. We will be using `~/kickoff-skeletons` for
the rest of this guide, but you are free to choose any path that suits you
best:

```bash
? Default skeleton repository ~/kickoff-skeletons
? Skeleton repository /home/johndoe/kickoff-skeletons does not exist, initialize it? [? for help] (Y/n)
```

This will create a new local skeleton repository and initializes it with a
default project skeleton so that we have something to work with:

```bash
? Skeleton repository /home/johndoe/kickoff-skeletons does not exist, initialize it? Yes
• creating skeleton repository      path=/home/johndoe/kickoff-skeletons
• creating skeleton directory       path=/home/johndoe/kickoff-skeletons/skeletons/default
• writing .kickoff.yaml             path=/home/johndoe/kickoff-skeletons/skeletons/default/.kickoff.yaml
• writing README.md.skel            path=/home/johndoe/kickoff-skeletons/skeletons/default/README.md.skel
```

Confirm the last two steps of the setup with `Y` to review and save the
configuration.

Kickoff will save its configuration in the user-level configuration directory
which depends on your OS. E.g. on Linux this will be either
`$XDG_CONFIG_HOME/kickoff/config.yaml` or `~/.config/kickoff/config.yaml`, on
MacOS this may be `$HOME/Library/Application Support/kickoff/config.yaml`.

If these paths do not suit you, you can learn how to change them in the
[configuration guide](configuration).

Now that the initial configuration is done, we are ready to create our first project.

## Creating your first project

We start off by verifying that Kickoff knows about the skeleton repository we
created in the previous section:

```bash
$ kickoff repository list

NAME    TYPE    PATH                    URL     REVISION
default local   ~/kickoff-skeletons     -       -
```

As you can see, we have a local repository named `default` which acts as a
source for our project skeletons.

Let's see what's in there:

```bash
$ kickoff skeleton list

REPONAME        NAME    PATH
default         default ~/kickoff-skeletons/skeletons/default
```

During the creating of our `default` repository it was also seeded with a
`default` project skeleton which we can use as a base to create our own. We
will learn about how to do that later. For now, we will just use it as is to
create our first project from a project skeleton.

Let's say, we want to create a new project at `~/projects/myproject` using the
`default` skeleton from the `default` repository. By running the following
command it will setup the new project from our skeleton:

```bash
$ kickoff project create default:default ~/projects/myproject

• creating project in /home/johndoe/projects/myproject
• creating directory                path=.
• rendering template                path.src=README.md.skel  path.target=README.md
• project creation complete
```

What happened here? Kickoff created the new directory `~/projects/myproject`
and copied all files and directories from the `default` skeleton into it,
evaluating eventual templates along the way. Noticed how `README.md.skel`
became `README.md`?

If we inspect it, we can see that the name of our new project was injected into
the README automatically:

```bash
$ cat ~/project/myproject/README.md

# myproject
...
```

That's it! We created our first project from an -- admittedly very poor --
project skeleton. In the [next section](skeletons/getting-started) you will
learn how to create your own project skeleton.

## Next steps

* [Creating project skeletons](skeletons/creating-skeletons): Learn more about
  how to create and share your own project skeletons.
* [Working with skeleton repositories](repositories): Using local and remote
  skeleton repositories.
* [Configuring Kickoff](configuration): Documentation of Kickoff's
  configuration options.
