---
title: Shell completion
permalink: /configuration/shell-completion
parent: Configuration
---

# Setting up shell completion
{: .no_toc }

To improve the user experience kickoff provides shell completion for `bash`,
`zsh`, `fish` and `powershell`. For detailed instructions about completion
setup consult the completion help:

```bash
$ kickoff completion --help
```

1. TOC
{:toc}

## Bash completion

Add to your `~/.bashrc` for bash completion:

```bash
source <(kickoff completion bash)
```

## Zsh completion

Add to your `~/.zshrc` for zsh completion:

```bash
source <(kickoff completion zsh)
```

## Fish completion

To load completions for every new session, run:

```bash
$ kickoff completion fish > ~/.config/fish/completions/kickoff.fish
```

## PowerShell completion

To load completions for every new session, run:

```bash
PS> kickoff completion powershell > kickoff.ps1
```

Afterwards source `kickoff.ps1` from your PowerShell profile.

## Next steps

* [Skeletons](/skeletons): Learn more about project skeletons.
