---
title: Inheritance
parent: Skeletons
nav_order: 5
---

{% include wip.md %}

# Skeleton inheritance
{: .no_toc }

1. TOC
{:toc}

Skeletons can inherit from other skeletons. Just add the `parent` configuration
to the `.kickoff.yaml` of the skeleton like this:

```yaml
parent:
  repositoryURL: https://github.com/martinohmann/kickoff-skeletons?revision=master
  skeletonName: my-parent-skeleton
```

If `repositoryURL` is omitted, the same repository as the one of the skeleton
is assumed. `repositoryURL` can be a remote URL or local path. Remote
repository urls can contain an optional `revision` query parameter which may
point to a commit, tag or branch. If omitted `master` is assumed.

## Next steps

* [Skeleton composition](composition): Creating projects from multiple project skeletons.
