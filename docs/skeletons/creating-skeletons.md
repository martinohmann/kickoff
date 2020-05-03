---
title: Creating skeletons
parent: Skeletons
nav_order: 2
---

# Creating project skeletons

In the [previous section](structure) you learned how project skeletons are
structured. With that information at hand you can go ahead and create your own.
Just create a new directory below the `skeletons` directory of your skeleton
repository and drop an empty `.kickoff.yaml` in there.

However, `kickoff` already provides a command for you that does just that and a
little more:

```bash
$ kickoff skeleton create ~/kickoff-skeletons/skeletons/myskeleton

• creating skeleton directory       path=/home/johndoe/kickoff-skeletons/skeletons/myskeleton
• writing .kickoff.yaml             path=/home/johndoe/kickoff-skeletons/skeletons/myskeleton/.kickoff.yaml
• writing README.md.skel            path=/home/johndoe/kickoff-skeletons/skeletons/myskeleton/README.md.skel
```

This will create a new skeleton called `myskeleton` and seeds it with a
documented `.kickoff.yaml` and an example `README.md.skel` template to get you
started.

You should see the newly created skeleton in the skeleton list now:

```bash
$ kickoff skeleton list

REPONAME        NAME                    PATH
default         default                 ~/kickoff-skeletons/skeletons/default
default         myskeleton              ~/kickoff-skeletons/skeletons/myskeleton
```

Happy templating!

## Next steps

* [Skeleton configuration](configuration): Learn more about
  the skeleton configuration file.
* [Templating](templating): Learn more about `.skel` templates and the usage of
  template variables within file and directory names.
