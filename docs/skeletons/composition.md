---
title: Composition
permalink: /skeletons/composition
parent: Skeletons
nav_order: 6
---

{% include wip.md %}

# Skeleton composition
{: .no_toc }

1. TOC
{:toc}

Projects can be created by composing multiple skeletons together. This is just
as simple as providing multiple skeletons instead of one after the project name
on project creation:

```bash
$ kickoff project create myproject repo:skeleton1 otherrepo:skeleton2 skeleton3
```

Note that the skeletons are merged left to right, so files and values from
skeletons on the right will override files and values of the same name from
other skeletons.

## Next steps

* [Working with skeleton repositories](/repositories): Using local and remote
  skeleton repositories.
