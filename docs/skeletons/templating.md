---
title: Templating
parent: Skeletons
nav_order: 4
---

# Templating
{: .no_toc }

This section provides documentation about kickoff's skeleton template feature.
You will learn how to create skeleton templates, which template functions are
available, and how to set template variables.

1. TOC
{:toc}

## Creating a template

Within a kickoff skeleton, any file with the `.skel` extension is treated as a
[Go template](https://golang.org/pkg/text/template/) and can be fully
templated.

When creating a project, kickoff will pass project and skeleton specific
variables to these `.skel` templates and then renders them. The rendered result
will be written to the target directory with the `.skel` extension stripped,
e.g. `README.md.skel` becomes `README.md`.

Since templates are just text files, the only thing that you need to do is to
make sure that your template files have the `.skel` extension, otherwise
kickoff will just treat them as normal files and do not run them through the
template rendering engine.

Below is a simply example of a skeleton template file called `README.md.skel`:

{% raw %}
```mustache
# {{.Project.Name}}

[![Build Status](https://travis-ci.org/{{.Project.Owner}}/{{.Project.Name}}.svg?branch=master)](https://travis-ci.org/{{.Project.Owner}}/{{.Project.Name}})

## Installation

To install `{{.Project.Name}}` execute `make install`.

{{ if .Values.allowContributions -}}
## Contributing

If you want to contribute, I'm happy to receive PRs!
{{ end -}}
```
{% endraw %}

Upon project creation, this gets rendered using the project information made
available in the `.Project` variable, and user-defined values in the `.Values`
variable. For example, if the project name is `myproject` and you configured
the default project owner to be `johndoe` while [initializing the kickoff
configuration](getting-started.html#initializing-the-kickoff-configuration),
this gets into the file `README.md` rendered as:

{% raw %}
```markdown
# myproject

[![Build Status](https://travis-ci.org/johndoe/myproject.svg?branch=master)](https://travis-ci.org/johndoe/myproject)

## Installation

To install `myproject` execute `make install`.
```
{% endraw %}

You might have noticed that the part that speaks about contributions was not
rendered into the final `README.md`. The reason for that is because it is
wrapped in a conditional. To make this work you need to set the
`allowContributions` value first. You will learn about this in the next
section.

If you are not familiar with the Go templating engine yet, you should make
youself comfortable with it. Here are some useful resources to get you started:

- [Go template documentation](https://golang.org/pkg/text/template/):
  Detailed documentation about the Go template syntax and features.
- [Sprig function documentation](https://masterminds.github.io/sprig/): Kickoff
  uses a template function library called "sprig" which comes with a wide
  variety of useful functions that are made available in templates.

## Accessing and setting template variables

Template variables are accessible via `.Values`. This variable contains the
merged result of the `values` from your local kickoff `config.yaml`, the
project skeleton's `values` from `.kickoff.yaml` and any variables that were
set using `--set` or `--values` during project creation.

{% raw %}
For example, the following snippets all set the variable that is then
accessible via `{{.Values.myVar}}` in templates:
{% endraw %}

```bash
# contents of .kickoff.yaml or ~/.config/kickoff/config.yaml
---
values:
  myVar: myValue

# on project creation via --set:
$ kickoff project create default ~/myproj --set myVar=myValue

# on project creation via --values:
$ kickoff project create default ~/myproj --values values.yaml

# where values.yaml contains myVar:
---
myVar: myValue
```

## Project template variables

Next to the user-defined `.Values`, kickoff makes a couple of variables
available to template files upon project creation:

| Variable                 | Description                                                                   |
| ---                      | ---                                                                           |
| `.Project.Host`          | The git host you specified during `kickoff init`, e.g. `github.com`           |
| `.Project.Owner`         | The project owner you specified, e.g. `martinohmann`                          |
| `.Project.Name`          | The name you specified when running `kickoff project create`                  |
| `.Project.License`       | The name of the license, if you picked one                                    |
| `.Project.Gitignore`     | Comma-separated list of gitignore templates, if provided                      |
| `.Project.URL`           | The URL to the project repo, e.g. `https://github.com/martinohmann/myproject` |
| `.Project.GoPackagePath` | The package path for go projects, e.g. `github.com/martinohmann/myproject`    |

## Template functions

As noted above, kickoff uses the [sprig template function
library](https://masterminds.github.io/sprig/). This provides the biggest chunk
of the available template functions.

In addition, there are also some built-in Go template functions which you can
find [listed here](https://golang.org/pkg/text/template/#hdr-Functions).

Last but not least, kickoff provides a couple of template functions that are
not provided by any of the two.

| Function        | Description                                                                                                                            |
| ---             | ---                                                                                                                                    |
| `toYAML`        | Converts its argument to a YAML string                                                                                                 |
| `mustToYAML`    | Converts its argument to a YAML string and fails if there were errors during marshalling                                               |
| `goPackageName` | Convenience function which creates a useful package name from a golang package path. E.g. `github.com/johndoe/my-pkg` becomes `mypkg` |

If feel that there are some useful functions missing, please feel free to open
an issue in the [kickoff project](https://github.com/martinohmann/kickoff) and
we'll see if it's worth adding.

## Templating file and directory names

Kickoff will try to resolve Go template variables in file and directory names.
E.g. a directory `{% raw %}cmd/{{.Project.Name}}{% endraw %}` will be resolved to `cmd/myproject` if
the project is named `myproject`.

It is also possible to put arbitrary files
into these directories. These will be moved to the correct place after the
directory name was resolved. You can take a look at [this example skeleton](https://github.com/martinohmann/kickoff-skeletons/tree/master/skeletons/golang/cli) in the [kickoff-skeletons](https://github.com/martinohmann/kickoff-skeletons) repository which makes use of this feature.

## Conditional inclusion of files

Sometimes it might be useful to only include files in a project based on the
value of some variable. Kickoff will skip the creation of a file in the project
directory when the following rules apply:

- The file is a `.skel` template.
- The file content has a length of zero bytes after template rendering.

To make a template conditionally render to zero bytes, you can create a `.skel`
template with a content like this:

{% raw %}
```mustache
{{- if .Values.includeFile }}
This will be rendered conditionally
{{ end -}}
```

**Note:** make sure you use `{{-` and `-}}` to trim whitespace surrounding your
conditional to ensure that the rendered result is zero bytes long when the
conditional evaluates to `false`.
{% endraw %}

In the `.kickoff.yaml` of the skeleton, set the `includeFile` value to `false`:

```yaml
values:
  includeFile: false
```

Now, the file will only be written to the project if `includeFile` is set to
`true` via the `--set` or `--values` as described in the [Accessing and setting
template variables section](#accessing-and-setting-template-variables).

**Note:** You can also force the inclusion of any empty file using the
`--allow-empty` flag upon project creation.

## Next steps

* [Skeleton inheritance](inheritance): Inheriting from a parent skeleton.
* [Skeleton composition](composition): Creating projects from multiple project skeletons.
