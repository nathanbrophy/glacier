---
title: glacier version
---

# glacier version    [ SDK ]

[ View source spec → ](../../../specs/0032-sdk.md#commands-version)
**Other commands:** [vibe](./vibe.md) [generate](./generate.md) [lint](./lint.md) [test](./test.md) [init](./init.md) [new](./new.md) [completions](./completions.md) [explain](./explain.md)

<!-- magpie:extract source=specs/0032-sdk.md section=commands subsection=version source-checksum=<TODO> -->
**Synopsis.** Print the Glacier SDK version. With `--check`, compare against the latest published release.

**Mental model.** `version` reads its own version (baked at build time via `-ldflags`; falls back to `runtime/debug.ReadBuildInfo` for `go run` builds), the Go toolchain version, and the OS/arch. `--check` adds a GitHub Releases lookup via `httpc`. The lookup is cached in `<UserCacheDir>/glacier/versioncheck.json` with a 24-hour TTL. Network failures degrade gracefully: stale cache is returned with a `(stale)` annotation; no cache available means `latest: unknown (offline)`. Exit 0 unless `--strict` opts into exit 68 on network failure.

**Flags.**

| Flag | Default | Description |
|---|---|---|
| `--check` | `false` | Contact GitHub Releases for the latest tag. |
| `--strict` | `false` | Make a network failure during `--check` exit 68. |
| `--json` | `false` | Emit a single JSON object instead of human-readable lines. |

**Output destination.** Version output writes to stdout. This means `glacier version --json | jq .latest` works correctly.

**Exit codes.** `0` success or offline with default behavior; `68` network failure when `--strict` is set; `130` SIGINT during network call.
<!-- /magpie:extract -->

## Try it

```asciinema
site/public/casts/version.cast
```

The cast shows `glacier version` followed by `glacier version --check`.

## Related commands

[explain](./explain.md) [completions](./completions.md)
