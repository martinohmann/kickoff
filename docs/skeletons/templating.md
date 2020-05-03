---
title: Templating
parent: Skeletons
nav_order: 4
---

{% include wip.md %}

# Templating
{: .no_toc }

Any file with the `.skel` extension is treated as a [Go
templates](https://golang.org/pkg/text/template/) and can be fully templated.
When creating a project the resolved `.skel` templates will be written to the
target directory with the `.skel` extension stripped, e.g. `README.md.skel`
becomes `README.md`.

1. TOC
{:toc}

## Template variables

Kickoff makes available a couple of variables to these templates:

| Variable                 | Description                                                                                                                                                                                         |
| ---                      | ---                                                                                                                                                                                                 |
| `.Project.Host`          | The git host you specified during `kickoff init`, e.g. `github.com`                                                                                                                                 |
| `.Project.Owner`         | The project owner you specified, e.g. `martinohmann`                                                                                                                                                |
| `.Project.Email`         | The project email you specified, e.g. `foo@bar.baz`                                                                                                                                                 |
| `.Project.Name`          | The name you specified when running `kickoff project create`                                                                                                                                        |
| `.Project.License`       | The name of the license, if you picked one                                                                                                                                                          |
| `.Project.Gitignore`     | Comma-separated list of gitignore templates, if provided                                                                                                                                            |
| `.Project.URL`           | The URL to the project repo, e.g. `https://github.com/martinohmann/myproject`                                                                                                                       |
| `.Project.GoPackagePath` | The package path for go projects, e.g. `github.com/martinohmann/myproject`                                                                                                                          |
| `.Project.Author`        | If an email is present, this will resolve to `project-owner <foo@bar.baz>`, otherwise just `owner`                                                                                                  |
| `.Values`                | The merged result of the `values` from your local kickoff `config.yaml`, the project skeleton's `values` from `.kickoff.yaml` and any variables that were set using `--set` during project creation |

## Templating file and directory names

Kickoff will try to resolve Go template variables in file and directory names.
E.g. a directory `{% raw %}cmd/{{.Project.Name}}{% endraw %}` will be resolved to `cmd/myproject` if
the project is named `myproject`.

It is also possible to put arbitrary files
into these directories. These will be moved to the correct place after the
directory name was resolved. You can check out this example skeleton which
makes use of this feature:
[kickoff-skeletons:golang/cli](https://github.com/martinohmann/kickoff-skeletons/tree/master/skeletons/golang/cli).

## Next steps

* [Skeleton inheritance](inheritance): Extending skeletons.
* [Skeleton composition](composition): Creating projects from multiple project skeletons.
