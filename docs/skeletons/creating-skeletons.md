---
title: Creating skeletons
permalink: /skeletons/creating-skeletons
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
$ kickoff skeleton create default myskeleton

✓ Created new skeleton myskeleton in repository default

You can inspect it by running: kickoff skeleton show default:myskeleton
```

This will create a new skeleton called `myskeleton` in the `default` and seeds
it with a documented `.kickoff.yaml` and an example `README.md.skel` template
to get you started.

You should see the newly created skeleton in the skeleton list now:

```bash
$ kickoff skeleton list

Repository  Name
default     default
default     myskeleton
```

You can also inspect inpect individual skeletons to see what's in there:

```bash
$ kickoff skeleton show myskeleton

Repository  default
Name        myskeleton
Path        ~/kickoff-skeletons/skeletons/myskeleton

Files               Values
myskeleton          null
└── README.md.skel
```

## Next steps

* [Skeleton configuration](configuration): Learn more about
  the skeleton configuration file.
* [Templating](templating): Learn more about `.skel` templates and the usage of
  template variables within file and directory names.
