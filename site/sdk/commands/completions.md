---
title: glacier completions
---

# glacier completions    [ SDK ]

[ View source spec → ](../../../specs/0032-sdk.md#commands-completions)
**Other commands:** [vibe](./vibe.md) [version](./version.md) [generate](./generate.md) [lint](./lint.md) [test](./test.md) [init](./init.md) [new](./new.md) [explain](./explain.md)

<!-- magpie:extract source=specs/0032-sdk.md section=commands subsection=completions source-checksum=<TODO> -->
**Synopsis.** Print a shell-completion script for the named shell.

**Mental model.** `completions <shell>` walks `cli.Default`'s registered command tree and emits a printf-friendly script for one of the four supported shells. The script is printed to stdout; the user redirects to the appropriate shell-completion location. No animation; no banner; no log records.

**Argument.**

```
glacier completions <shell>
```

`<shell>` is required. Values: `bash`, `zsh`, `fish`, `pwsh`.

**Exit codes.** `0` success; `2` unknown shell; `1` stdout write failure.
<!-- /magpie:extract -->

## Try it

```
$ glacier completions bash
# bash completion for glacier                    -*- shell-script -*-
_glacier_completions() {
    local cur prev words cword
    _init_completion || return
    case "${prev}" in
        glacier)
            COMPREPLY=( $(compgen -W "vibe version generate lint test init new completions explain" -- "${cur}") )
            return
            ;;
        completions)
            COMPREPLY=( $(compgen -W "bash zsh fish pwsh" -- "${cur}") )
            return
            ;;
        # ...
    esac
}
complete -F _glacier_completions glacier
```

## Shell-specific install

| Shell | Install location |
|---|---|
| bash | `~/.local/share/bash-completion/completions/glacier` |
| zsh | `${fpath[1]}/_glacier` (then run `compinit`) |
| fish | `~/.config/fish/completions/glacier.fish` |
| PowerShell | Append to `$PROFILE` |

See [Install](../install.md#shell-completions) for the full install commands.

## Related commands

[version](./version.md) [explain](./explain.md)
