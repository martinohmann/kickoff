---
title: Structure
permalink: /skeletons/structure
parent: Skeletons
nav_order: 1
---

# Skeleton structure
{: .no_toc }

A project skeleton is just a simple directory inside the `skeletons` directory
of a skeleton repository. As a minimum, the skeleton directory needs to contain
a `.kickoff.yaml` file which may be empty to be recognized by `kickoff`.

The `.kickoff.yaml` file can contain some metadata and additional configuration
for the skeleton which we will learn more about further down the page.


1. TOC
{:toc}

## Simple skeletons

A simple skeleton named `myskeleton` inside a skeleton repository at
`~/kickoff-skeletons` may look like this:

{% raw %}
```
~/kickoff-skeletons
└── skeletons
    └── myskeleton
        ├── .kickoff.yaml
        ├── README.md.skel
        └── somefile.txt
```
{% endraw %}

You might have spotted the file with the `.skel` extension. This is a special
skeleton template file that gets evaluated during project creation and can
contain variables for project specific information or even user-defined values.
The [templating guide](templating) has more detailed information
about that.

## Grouping skeletons into topics

Sometimes it is a good idea to group thing together based on a specific topic
to keep things organized. For example, you might want to group project
skeletons by programming language, business unit or concept.

Kickoff repositories support nesting directories below the `skeletons` dir.

{% raw %}
```
~/kickoff-skeletons
└── skeletons
    └── topic
        ├── myskeleton
        │   └── .kickoff.yaml
        └── otherskeleton
            └── .kickoff.yaml
```
{% endraw %}

The example above shows a repository that contains two skeletons named
`topic/myskeleton` and `topic/otherskeleton`. Please note that the `topic`
directory **must not** contain a `.kickoff.yaml` file itself as nesting
skeletons is not supported.

## Advanced skeletons with file and directory name templating

Sometimes, file templates are just not flexible enough and we also want to
dynamically name certain project files or directories based on some project
specific or user-defined values.

This can be achieved by templating the filenames itself:

{% raw %}
```
.~/kickoff-skeletons
└── skeletons
    └── someskeleton
        ├── .kickoff.yaml
        ├── README.md.skel
        ├── somedir
        │   └── {{.Values.somedirname}}
        │       ├── someotherfile.json
        │       ├── someothertemplate.txt.skel
        │       └── some-{{.Value.someval}}.json.skel
        ├── somefile.json
        ├── {{.Project.Name}}.conf
        ├── sometemplate.yaml.skel
        └── {{.Values.filename}}.txt
```

Here, file and directory names that contain template variables (e.g.
`{{.Values.someval}}`) will be built upon project creation. The full details of
that are covered in the [templating guide](templating).

{% endraw %}

## Next steps

* [Creating project skeletons](creating-skeletons): Learn more about
  how to create your own project skeletons.
* [Skeleton configuration](configuration): Learn more about
  the skeleton configuration file.
* [Templating](templating): Learn more about `.skel` templates and the usage of
  template variables within file and directory names.
