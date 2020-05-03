---
title: Configuration
nav_order: 4
has_children: true
---

{% include wip.md %}

# Configuration
{: .no_toc }

Kickoff will save its configuration in the user-level configuration directory
which depends on your OS. E.g. on Linux this will be either
`$XDG_CONFIG_HOME/kickoff/config.yaml` or `~/.config/kickoff/config.yaml`, on
MacOS this may be `$HOME/Library/Application Support/kickoff/config.yaml`.

If that does not suit you, use the `KICKOFF_CONFIG` [environment
variable](/configuration/environment-variables) to override this.

Alternatively, you can also pass the configuration file path via the `--config`
flag to kickoff commands that make use of the config.

1. TOC
{:toc}

## Editing and inspecting the configuration

To edit the configuration file, run:

```bash
$ kickoff config edit
```

This will launch an editor where you can make your changes. The config file is
saved once you close the editor. Kickoff will try to use the editor you
configured in the `EDITOR` environment variable if present. If you do not like
the default editor, you can override it with the `KICKOFF_EDITOR` [environment
variable](/configuration/environment-variables).

If you just want to quickly inspect the configuration file, you can run the
following command to print it:

```bash
$ kickoff config show
```

## Configuration file structure

The kickoff configuration file structure looks like this:

```yaml
---
project:
  email: johndoe@example.com
  gitignore: none
  host: github.com
  license: none
  owner: johndoe
repositories:
  default: ~/kickoff-skeletons
values: {}
```

The `email`, `owner` and `host` configuration fields are useful to be able to
create project specific links or copyright notices in skeleton templates.

## Configuring a default `license`

If you want to set a default license that is used for every new project (unless
overridden), you can set that in the `license` field of the `project`
configuration. Leaving the field empty or setting it to `none` disables the
inclusion of a default `LICENSE` file into new projects.

You can find our of the supported license names by running:

```bash
$ kickoff license list
```

Showing the license text of a specific license works like this:

```bash
$ kickoff license show mit
```

## Configuring default `gitignore` templates

The `gitignore` field of the `project` configuration lets you specify a
comma-separated list of `.gitignore` templates that should be added into the
`.gitignore` file of a new project. This can be overridding on project
creation. Leaving this field empty or setting it to `none` will disable the
inclusion of a default `.gitignore` file.

```bash
$ kickoff gitignore list
```

You can test out how the generated `.gitignore` looks like by showing the
gitignore templates you want to use all together. For example:

```bash
$ kickoff gitignore show go,hugo
```

## Configuring default `values`

In the `values` map you can configure default values that get merged on top of
the values of every skeleton that a project is created with.

Refer to the documentation of [skeleton default
values](/skeletons/configuration#configuring-default-values) for more
information.

**Caution**: Since values may have a different meaning in different skeletons,
configuring global value defaults can cause project creation to fail due to
value type errors.
