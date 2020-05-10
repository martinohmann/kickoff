---
title: Repositories
permalink: /repositories
nav_order: 7
has_children: false
---

{% include wip.md %}

# Using skeleton repositories

Kickoff supports local and remote skeleton repositories. If you want, you can
use the repository that come along with `kickoff`. Head over to the
[kickoff-skeletons](https://github.com/martinohmann/kickoff-skeletons)
repository for ready-to-use skeletons and to get some inspiration to create
your own.

You can add the `kickoff-skeletons` repository to your config to directly
create projects from the available skeletons:

```bash
$ kickoff repository add kickoff-skeletons https://github.com/martinohmann/kickoff-skeletons
```

## Local skeleton repositories

Kickoff supports local repositories which do not necessarily need to be git
repos. If you did not create a local repository via `kickoff init`, you can
create one like this:

```bash
$ kickoff repository create ~/path/to/new/repo
$ kickoff repository add myrepo ~/path/to/new/repo
```

The `kickoff repository create` command will create a new repository which
already contains a minimal `default` skeleton with a commented `.kickoff.yaml`
file and a `README.md.skel` skeleton to get you started. You can delete it or
customize it to your needs.

You can verify that your local repository was correctly created and added by
listing the available kickoff repositories:

```bash
$ kickoff repository list
```

## Remote skeleton repositories

Add a remote skeleton repository and create a new project:

```bash
$ kickoff repository add myremoterepo https://github.com/myuser/myskeletonrepo?revision=v1.0.0
$ kickoff repository list
$ kickoff project create myremoterepo:myskeleton ~/path/to/my/new/project
```

Remote repository urls can contain an optional `revision` query parameter which
may point to a commit, tag or branch. If omitted `master` is assumed.
