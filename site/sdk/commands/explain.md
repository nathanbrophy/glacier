---
title: glacier explain
---

# glacier explain

**Synopsis.** Print a boxed explanation for a Glacier marker, exit code, or config key.

**Other commands:** [vibe](./vibe.md) [version](./version.md) [generate](./generate.md) [lint](./lint.md) [test](./test.md) [init](./init.md) [new](./new.md) [completions](./completions.md)

## Argument

```
glacier explain <topic>
glacier explain --list
```

`<topic>` is one of:

- a marker name: `+glacier:command`, `+glacier:mock`, `+glacier:positional`, etc.
- an exit code: `64`, `65`, `66`, `67`, `68`, `69`, `70`, `130`, `143`
- a config key: `github.repo`, `versioncheck.cache_ttl`, `versioncheck.enabled`, `versioncheck.strict`, `banner.show_on_help`

## Flags

| Flag | Default | Description |
|---|---|---|
| `--list` | `false` | Print all available topics grouped by category and exit. |

## Examples

Explain an exit code:

```sh
glacier explain 66
```

Explain a marker:

```sh
glacier explain +glacier:command
```

Explain a config key:

```sh
glacier explain versioncheck.cache_ttl
```

List all available topics:

```sh
glacier explain --list
```

Look up without typing the full slug (did-you-mean kicks in for close matches):

```sh
glacier explain verison  # suggests "versioncheck.enabled" or exit codes
```

## What it does under the hood

`explain` reads from an embedded `embed.FS` of pre-rendered topic files at `cmd/glacier/internal/explain/topics/<slug>.md`. The files are generated at build time by `cmd/glacier/internal/explaingen`. When the topic is found, it is rendered inside a `term.Box` with rounded corners; the title kaomoji reflects the category (`ʕ•_•ʔ` for markers, `ʕ× ×ʔ` for exit codes, `ʕ⌐■-■ʔ` for config keys). When the topic is not found, Levenshtein distance up to 2 produces a "did you mean?" suggestion. `--list` prints all topics to stdout grouped by category with no box rendering.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Topic found and printed, or `--list` completed |
| 1 | stdout write failure |
| 2 | Unknown topic (with did-you-mean hint when available) |

## See also

- [`glacier version`](./version.md) - uses the explain topic for exit 68
- [`glacier lint`](./lint.md) - uses the explain topic for exit 65
- [`glacier test`](./test.md) - uses the explain topic for exit 66
- [`glacier completions`](./completions.md)
