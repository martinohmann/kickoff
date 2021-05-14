---
title: Home
nav_order: 1
---

# kickoff -- A project bootstrapping tool

Start new projects from reusable skeleton directories. Use community project
skeletons or create your own. No more need to copy & paste initial boilerplate
like Makefiles, CI configuration or language specific configuration files from
existing projects to a new one.
{: .fs-6 }

[![asciicast](https://asciinema.org/a/414074.svg)](https://asciinema.org/a/414074)

[Getting started](getting-started){: .btn .btn-primary .fs-5 .mb-4 .mb-md-0 .mr-2 }
[View on GitHub](https://github.com/martinohmann/kickoff){: .btn .fs-5 .mb-4 .mb-md-0 .mr-2 }

<hr>

## Features

- Templating of filenames, directory names and file contents via [Go
  templates](https://golang.org/pkg/text/template/) and
  [Sprig](http://masterminds.github.io/sprig/).
- Extensible by allowing users to pass arbitrary values to templates via config
  files or CLI flags.
- Automatically populate LICENSE file with an open source license obtained from
  the [GitHub Licenses API](https://docs.github.com/en/rest/reference/licenses).
- Automatically add a .gitignore created from templates obtained from
  [GitHub Gitignores API](https://docs.github.com/en/rest/reference/gitignore).
- Set local author, repository and skeleton defaults using custom config file.
- Skeleton composition: projects can be created by composing multiple skeletons
  together.
