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

[![asciicast](https://asciinema.org/a/T53cAY9Uitt4I8XQT5rWPKDxk.svg)](https://asciinema.org/a/T53cAY9Uitt4I8XQT5rWPKDxk)

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
  the [GitHub Licenses API](https://developer.github.com/v3/licenses/).
- Automatically add a .gitignore created from templates obtained from
  [gitignore.io](https://gitignore.io).
- Set local author, repository and skeleton defaults using custom config file.
- Dry run for project creation.
- Skeleton inheritance: skeletons can inherit files and values from an optional
  parent skeleton.
- Skeleton composition: projects can be created by composing multiple skeletons
  together. This is similar to inheritance but allows for way more flexible use
  of skeletons.
