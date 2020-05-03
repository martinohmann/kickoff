---
title: Skeletons
nav_order: 5
has_children: true
---

# Project skeletons

Project skeletons are blueprints that can be used to create new projects. They
can be fully templated and composed, or inherit from other skeletons.

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
$ kickoff skeleton create ~/kickoff-skeletons/skeletons/myskeleton
```

However, there is more to it. Check out the [Creating project
skeletons](/skeletons/creating-skeletons) section to find out.

### Listing skeletons

This is how you list all skeletons that are available in the skeleton
repositories you configured:

```bash
$ kickoff skeleton list

REPONAME        NAME                    PATH
default         default                 ~/kickoff-skeletons/skeletons/default
default         myskeleton              ~/kickoff-skeletons/skeletons/myskeleton
```

To learn more about the working with repositories, head over to the
[repositories documentation](/repositories).

### Inspecting skeletons

In addition to listing skeletons, they can also be inspected individual:

```bash
$ kickoff skeleton show myskeleton

Name            myskeleton
Path            ~/kickoff-skeletons/skeletons/myskeleton
Description     -
Files           myskeleton
                └── README.md.skel
Parent          -
Values          travis:
                  enabled: false
```

**Hint**: try out the `--output yaml` flag to get even more information.
