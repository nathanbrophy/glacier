---
title: glacier version
---

# glacier version

**Synopsis.** Print the Glacier SDK version, Go toolchain, and OS/arch. With `--check`, compare against the latest published release.

**Other commands:** [vibe](./vibe.md) [generate](./generate.md) [lint](./lint.md) [test](./test.md) [init](./init.md) [new](./new.md) [completions](./completions.md) [explain](./explain.md)

## Flags

| Flag | Default | Description |
|---|---|---|
| `--check` | `false` | Fetch the latest release from GitHub and compare against the running version. |
| `--strict` | `false` | When combined with `--check`, exit 68 if the GitHub endpoint is unreachable instead of degrading gracefully. |
| `--json` | `false` | Emit a single JSON object to stdout instead of human-readable lines. |

## Examples

Print the current version:

```sh
glacier version
```

Check whether a newer release is available:

```sh
glacier version --check
```

Machine-readable output for scripting:

```sh
glacier version --json
```

Fail the script if the version check cannot reach GitHub:

```sh
glacier version --check --strict
```

Parse the latest tag with jq:

```sh
glacier version --check --json | jq -r '.latest.tag'
```

## What it does under the hood

`version` reads its own version string from a value baked in at build time via `-ldflags`; when that is empty or `"dev"` it falls back to `runtime/debug.ReadBuildInfo`. With `--check`, it looks up the latest release via `cmd/glacier/internal/ghreleases` using `httpc.Default`. The lookup result is cached in a layered cache (in-memory primary backed by `<UserCacheDir>/glacier/`) with a 24-hour TTL configured by `versioncheck.cache_ttl`. A network failure without `--strict` prints `latest: unknown (offline)` and exits 0. All version output writes to stdout so `glacier version --json | jq .latest` works correctly.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Success, or offline without `--strict` |
| 68 | Network failure during `--check` when `--strict` is set |
| 130 | SIGINT during the network call |

## See also

- [`glacier explain versioncheck.cache_ttl`](./explain.md) - cache TTL config key
- [`glacier completions`](./completions.md)
