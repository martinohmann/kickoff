---
title: Environment variables
permalink: /configuration/environment-variables
parent: Configuration
---

# Environment variables

The following environment variables can be used to configure kickoff:

| Name                | Description                                                                                          |
| ---                 | ---                                                                                                  |
| `KICKOFF_CONFIG`    | Override path to the kickoff config.                                                                 |
| `KICKOFF_EDITOR`    | Editor used by `kickoff config edit`. If unset, `EDITOR` environment will be used. Fallback is `vi`. |
| `KICKOFF_LOG_LEVEL` | Sets the kickoff log level. Can be overridden with the `--log-level` flag                            |

## Next steps

* [Shell completion](shell-completion): Setup shell completion to enhance your user experience.
