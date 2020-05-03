---
title: Configuration
parent: Skeletons
nav_order: 3
---

# Skeleton configuration
{: .no_toc }

Skeletons can carry additional configuration in a mandatory `.kickoff.yaml`
file. The following sections explain the structure of the file and the purpose
of the fields that can be configured in it.

1. TOC
{:toc}

## The `.kickoff.yaml` file

{% raw %}
```
~/kickoff-skeletons
└── skeletons
    └── myskeleton
        └── .kickoff.yaml
```
{% endraw %}

The `.kickoff.yaml` needs to be present within a skeleton
directory for kickoff to be able to identify it as such.

The file has the following special properties:

* Acts as a marker to identify the root of a project skeleton.
* May contain metadata about a skeleton like description or parent.
* May contain defaults for values used within templates.
* Defining metadata and value defaults is totally optional, thus a valid
  `.kickoff.yaml` can also be empty.

### General file structure

The following code block shows the contents of an example `.kickoff.yaml`:

```yaml
---
description: |
  Some optional description of the skeleton that might be helpful to users.

  Upon project creation you may want to pass `--set travis.enabled=true` if you
  want to enable travis-ci for your project!
parent:
  repositoryURL: https://github.com/someuser/kickoff-skeletons?revision=v1.0
  skeletonName: javascript/react-project
values:
  myVar: 'myValue'
  travis:
    enabled: false
```

The individual config options will be explained below.

### The `description` field

This is the place where project skeleton creators may put some useful
description of the skeleton and its usage. E.g. it can be used to explain the
meaning of the skeleton values and give some examples of how to customize them.

Users can view the `description` of a skeleton by inspecting the skeleton:

```bash
$ kickoff skeleton show <name-of-the-skeleton>
```

### Configuring a `parent` skeleton

Sometimes you might want to extend an existing skeleton instead of duplicating
it. This is where the `parent` field comes into play. It lets you specify a
parent skeleton which can either reside in a remote skeleton repository or in
the same repository as the skeleton itself.

Check out the [inheritance documentation](inheritance) for details.


### Configuring default `values`

You can make use of user-defined values in your project skeletons which are
make available below the `.Values` variable in templates. The values can have
whatever type or structure you like. You can even nest them.

It is good practice to set sane defaults for values you are using in your
skeleton so that users do not run into problems or weird behaviour that comes
along with undefined variables.

Users can override these values upon project creation using the `--set` or
`--values` CLI flags.

For example, the following configuration sets the default for the template
variable `.Values.myVar` to `myValue`:

```yaml
values:
  myVar: 'myValue'
```

Users can then override it via `--set myVar=someOtherValue` or override
multiple values from file via `--values`.

## Next steps

* [Templating](templating): Learn more about `.skel` templates and the usage of
  template variables within file and directory names.
* [Skeleton inheritance](inheritance): Inheriting from a parent skeleton.
* [Skeleton composition](composition): Creating projects from multiple project skeletons.
