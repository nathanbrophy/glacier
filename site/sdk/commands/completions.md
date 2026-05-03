---
title: glacier completions
---

# glacier completions

**Synopsis.** Print a shell-completion script for bash, zsh, fish, or PowerShell.

**Other commands:** [vibe](./vibe.md) [version](./version.md) [generate](./generate.md) [lint](./lint.md) [test](./test.md) [init](./init.md) [new](./new.md) [explain](./explain.md)

## Argument

```
glacier completions <shell>
```

`<shell>` is required. Accepted values: `bash`, `zsh`, `fish`, `pwsh`.

## Flags

| Flag | Default | Description |
|---|---|---|
| `<shell>` | (required) | Target shell. Values: `bash`, `zsh`, `fish`, `pwsh`. |

## Examples

Print the bash completion script to stdout:

```sh
glacier completions bash
```

Install permanently for bash:

```sh
glacier completions bash > ~/.local/share/bash-completion/completions/glacier
```

Load for the current bash session only:

```sh
source <(glacier completions bash)
```

Install for zsh:

```sh
glacier completions zsh > "${fpath[1]}/_glacier"
compinit
```

Install for fish:

```sh
glacier completions fish > ~/.config/fish/completions/glacier.fish
```

Install for PowerShell:

```powershell
glacier completions pwsh >> $PROFILE
```

## Shell install locations

| Shell | Install location |
|---|---|
| bash | `~/.local/share/bash-completion/completions/glacier` |
| zsh | `${fpath[1]}/_glacier` (run `compinit` after) |
| fish | `~/.config/fish/completions/glacier.fish` |
| PowerShell | Append to `$PROFILE` |

## What it does under the hood

`completions` looks up the requested shell in `cmd/glacier/internal/completions` and writes the pre-generated script to stdout. No animation, no banner, no log records. The script is generated once at build time from the registered command tree, so it covers all nine verbs and their flags. Redirecting stdout to the appropriate file is all that is needed for permanent installation.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Success |
| 1 | stdout write failure |
| 2 | Unknown shell name |

## See also

- [Install - Shell completions](../install.md#shell-completions) - per-platform install steps
- [`glacier version`](./version.md)
- [`glacier explain`](./explain.md)
