---
title: glacier lint
---

# glacier lint    [ SDK ]

[ View source spec → ](../../../specs/0032-sdk.md#commands-lint)
**Other commands:** [vibe](./vibe.md) [version](./version.md) [generate](./generate.md) [test](./test.md) [init](./init.md) [new](./new.md) [completions](./completions.md) [explain](./explain.md)

<!-- magpie:extract source=specs/0032-sdk.md section=commands subsection=lint source-checksum=<TODO> -->
**Synopsis.** Run gofmt, go vet, staticcheck, and Glacier-specific lints.

**Mental model.** `lint` runs an in-process suite for the cheap lints (gofmt via `go/format`, go vet via `golang.org/x/tools/go/analysis`, plus the six Glacier-specific lints) and shells out to `staticcheck` if it's on PATH. A content-hash cache at `<repo>/.glacier/lint-cache.json` reuses results across runs. Findings are grouped by severity (errors first), then by file. On capable terminals each `file:line` is an OSC-8 hyperlink. `--fix` auto-fixes gofmt + no-em-dash + marker normalization; other lints get a manual-fix hint. Exit 65 on findings >= severity threshold; exit 70 on subprocess failure.

**Flags.**

| Flag | Default | Description |
|---|---|---|
| `[patterns]` | `./...` | go/packages patterns to scan. |
| `--fix` | `false` | Apply auto-fixable lint fixes in place. |
| `--severity` | `warning` | Minimum severity to report. Values: `error`, `warning`, `info`. |
| `--format` | `text` | Output format. Values: `text`, `json`, `sarif`. |
| `--no-cache` | `false` | Ignore the lint result cache. |

**Glacier-specific lints.**

| Name | Severity | Description |
|---|---|---|
| `exported-doc-comment` | warning | Every exported symbol has a doc comment starting with the symbol name. |
| `package-example-test` | warning | Every package has at least one `Example*` function. |
| `panic-in-library` | error | No `panic(...)` outside `_test.go` in non-`cmd/` packages. |
| `no-em-dash` | error | No U+2014 in any `.go`, `.md`, `.txt` file. |
| `library-error-register` | error | Every exported `*Error` type's `Error()` matches `^[a-z][^.]*$`. |
| `naked-any` | warning (opt-in) | Type-parameter constraint of `any` where a more specific constraint would do. |

**Exit codes.** `0` clean; `2` bad pattern; `65` findings at or above severity threshold; `70` subprocess failure.
<!-- /magpie:extract -->

## Try it

```
$ glacier lint ./...
ʕ•ᴥ•ʔ glacier lint ./...
ʕ× ×ʔ ERRORS
  cli/app.go:142  panic-in-library     panic in non-test, non-cmd package: panic("unreachable")
  cli/gen/parse.go:67  no-em-dash      em-dash character (U+2014) found

ʕ◉_◉ʔ WARNINGS
  cli/app.go:88   exported-doc-comment  exported func App.Lookup has no doc comment
  conf/conf.go:51 package-example-test  package conf has no Example* function

ʕ•ᴥ•ʔ 3 findings: 2 errors, 1 warning.
exit 65
```

## Related commands

[generate](./generate.md) [test](./test.md) [explain](./explain.md)
