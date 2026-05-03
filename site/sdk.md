---
title: Glacier SDK
aside: false
---

# Glacier SDK

<div class="glacier-sprite-accent">
  <MascotSprite state="confident" :size="96" />
</div>

The `glacier` CLI is the public face of the Glacier framework: nine commands that scaffold projects, generate code, lint, test, animate, and answer reference questions. Color is on by default. Help is real. Every command is built on the framework it ships, so the binary doubles as the framework's longest-running integration test.

::: tip Try it now
```sh
go install github.com/nathanbrophy/glacier/cmd/glacier@latest
glacier --help
```
:::

## What ships

The SDK is nine commands, grouped by what you do with them.

<table class="sdk-cmd-table">
<thead><tr><th>Group</th><th>Command</th><th>What it does</th></tr></thead>
<tbody>
<tr><td rowspan="2"><strong>CREATE</strong></td>
    <td><a href="/sdk/commands/init">glacier init</a></td>
    <td>Scaffold a new Glacier project.</td></tr>
<tr><td><a href="/sdk/commands/new">glacier new</a></td>
    <td>Add a package, command, or option to an existing project.</td></tr>
<tr><td rowspan="3"><strong>DEVELOP</strong></td>
    <td><a href="/sdk/commands/generate">glacier generate</a></td>
    <td>Run all registered code generators (cli, mock, httpmock).</td></tr>
<tr><td><a href="/sdk/commands/lint">glacier lint</a></td>
    <td>gofmt + go vet + staticcheck + 6 Glacier-specific lints.</td></tr>
<tr><td><a href="/sdk/commands/test">glacier test</a></td>
    <td>Wraps go test with summary, bench baseline, JUnit/SARIF.</td></tr>
<tr><td rowspan="2"><strong>INSPECT</strong></td>
    <td><a href="/sdk/commands/explain">glacier explain</a></td>
    <td>Reference for markers, exit codes, config keys.</td></tr>
<tr><td><a href="/sdk/commands/version">glacier version</a></td>
    <td>Version info; <code>--check</code> fetches latest.</td></tr>
<tr><td rowspan="2"><strong>UTILITY</strong></td>
    <td><a href="/sdk/commands/completions">glacier completions</a></td>
    <td>Print a shell completion script.</td></tr>
<tr><td><a href="/sdk/commands/vibe">glacier vibe</a></td>
    <td>Animated bear + tip rotation; ambient framework reference.</td></tr>
</tbody>
</table>

## A real session

Every screencap below was captured from the live binary. The asciinema cast scripts live at `cmd/glacier/docs/casts/` and the rendered SVGs are committed under `site/public/casts/` so they cannot drift from reality.

### `glacier --help`

The top-level help page is the format-A renderer from spec 0032 D-S39: full block-bear plus wordmark gradient, then commands grouped by purpose, then the global flags table, then a footer pointing at per-command help and `glacier explain`.

<div class="cast-frame">
  <a href="/casts/help.cast" rel="noopener">
    <img src="/casts/help.svg" alt="glacier --help (animated)" loading="lazy" />
  </a>
  <div class="cast-caption">Recorded with <code>asciinema rec</code>; rendered via <code>agg</code> at site-build time.</div>
</div>

```text
                                  ██████╗ ██╗      █████╗  ██████╗██╗███████╗██████╗
    ▟▀▙   ▟▀▙                    ██╔════╝ ██║     ██╔══██╗██╔════╝██║██╔════╝██╔══██╗
   ▟████████▙                    ██║  ███╗██║     ███████║██║     ██║█████╗  ██████╔╝
   █ ●  ▼  ● █                   ██║   ██║██║     ██╔══██║██║     ██║██╔══╝  ██╔══██╗
    ▀▀▀▀▀▀▀▀▀                    ╚██████╔╝███████╗██║  ██║╚██████╗██║███████╗██║  ██║
      ʕ•ᴥ•ʔ

USAGE
  glacier [global flags] <command> [command flags] [args]

CREATE
  init  Scaffold a new Glacier project.
  new   Add a package, command, or option to an existing project.

DEVELOP
  generate  Run all registered code generators (cli, mock, httpmock).
  lint      Run gofmt + go vet + staticcheck + 6 Glacier-specific lints.
  test      Run go test with a streaming summary, bench baseline, JUnit/SARIF.

INSPECT
  explain  Show reference for a marker, exit code, or config key.
  version  Print version info; --check fetches the latest release.

UTILITY
  completions  Print a shell completion script (bash, zsh, fish, pwsh).
  vibe         Animated bear + tip rotation; ambient framework reference.

GLOBAL FLAGS
  -q, --quiet         Lower log level to Warn; suppress animations.
  -V, --verbose       Raise log level to Debug.
      --no-color      Disable ANSI color output.
      --force-color   Force ANSI color even when output is not a TTY.
      --no-banner     Suppress the banner on --help.
      --otelendpoint  OTEL_EXPORTER_OTLP_ENDPOINT override.

Run glacier <command> --help for command-specific help.
Run glacier explain <topic> for marker, exit-code, or config-key reference.
```

The wordmark gradient and command-name highlighting are 24-bit ANSI in real terminals; the static rendering above is plain text.

### `glacier vibe`

A meditative wordmark animation with the polar bear cycling through expressions and a tip rotation drawn from the framework's own documentation. Run it when you want the SDK to remind you that less plumbing means more Go.

<div class="cast-frame">
  <a href="/casts/vibe.cast" rel="noopener">
    <img src="/casts/vibe.svg" alt="glacier vibe (animated)" loading="lazy" />
  </a>
  <div class="cast-caption">10s loop; bear cycles every 3s, wordmark gradient shimmers per 100ms tick.</div>
</div>

### `glacier test ./...`

The test wrapper runs `go test -json` and renders a colored streaming panel during the run, then a per-metric summary box (pass green, fail red, skip yellow, package names blue, test names magenta) and an aligned bench results box when `--bench` is set.

```text
ʕ•ᴥ•ʔ glacier test
ʕ⌐■-■ʔ ./cache/...   (0.412s)
ʕ⌐■-■ʔ ./cli/...     (0.628s)
ʕ⌐■-■ʔ ./term/...    (0.554s)

╭─ glacier test summary  12:50:37 ──────────────────────────────╮
│  packages: 42  tests: 1757  pass: 1747  fail: 0  skip: 10    │
│  wall: 20.9s                                                  │
╰───────────────────────────────────────────────────────────────╯
ʕ⌐■-■ʔ that went well.
```

### `glacier test --bench=. -benchmem ./...`

The bench summary is a colored, column-aligned table of every measured benchmark. When a baseline exists at `.glacier/bench-baseline.json`, an extra column shows the percentage delta vs baseline (green = faster, red = regression, yellow = slight slowdown).

```text
╭─ bench results  60 benchmark(s) ─────────────────────────────────╮
│                                                                  │
│  name                          ns/op     B/op   allocs/op        │
│  ─────────────────────────────────────────────────────────       │
│  BenchmarkMemHit-24            13.1 ns      0           0        │
│  BenchmarkStyleRender-24        2.3 ns      0           0        │
│  BenchmarkBox-24                1.01 µs  1320          18        │
│  BenchmarkEqualLargeMap-24      102.3 µs 56664        5005       │
│                                                                  │
│  baseline: .glacier/bench-baseline.json                          │
╰──────────────────────────────────────────────────────────────────╯
```

### `glacier explain exit:66`

The reference command for the SDK's exit codes, marker grammar, and config keys. Each topic is sourced from the spec, generated at build time, and shipped via `embed.FS` so a release binary always carries its own documentation.

```text
ʕ•ᴥ•ʔ glacier explain
╭─ ʕ× ×ʔ  Exit code 66: tests failed ──────────────────────╮
│                                                          │
│  glacier test reported one or more test failures, or a   │
│  benchmark regressed by more than 5% vs the stored       │
│  baseline.                                               │
│                                                          │
│  For benchmark regressions, run                          │
│  glacier test --update-baseline to accept the new        │
│  performance level.                                      │
│                                                          │
│  See also: exit:65                                       │
╰──────────────────────────────────────────────────────────╯
```

## Color, by default

Color is on for every command in every terminal that supports ANSI. The decision is global and gated by a single function (`term.ShouldColor`), so the kaomoji status lines, banner gradient, box borders, format-A help, lint findings, test summary, and bench results are all colored when allowed and all plain when disabled.

```sh
# Color is on by default, no flag needed.
glacier version

# Toggle off by flag or env var.
glacier --no-color version
NO_COLOR=1 glacier version

# Force color even when piping (for less -R or capturing logs).
glacier --force-color test ./... | less -R
GLACIER_FORCE_COLOR=1 glacier test ./...
```

## Built on the framework

Every byte of the SDK is built on the Glacier framework's own libraries. That is the whole point: if a package is good enough to ship, the SDK uses it as a daily-driver. If a package isn't ergonomic enough for the SDK, that surfaces as a paper cut against the framework, not as private SDK plumbing.

| SDK feature | Framework package |
|---|---|
| Command dispatch + flag parsing | `cli` |
| Help format-A + format-B rendering | `cli` + `term` |
| Banner gradient, box borders, animator | `term` |
| Kaomoji status lines | `cmd/glacier/internal/report` (uses `term.ShouldColor`) |
| Logging via `log.NewHandler` | `log` |
| OTEL initialization (opt-in) | `obs` |
| Generators (cli, mock, httpmock) | `cli/gen`, `mock/gen`, `httpmock/gen` |
| Configuration loading | `conf` |
| Cache for `version --check` | `cache` (spec 0033) |
| HTTP client for GitHub Releases | `httpc` |
| Testable HTTP via `httpmock` | `httpmock` |
| Subcommand parallelism in `generate` | `concur` |
| Lint findings stream | `fluent` |
| Atomic file writes | `internal/safefile` |
| Cross-platform flock | `internal/lockfile` |

The `TestGlacierEverywhere` canary in `cmd/glacier/commands/glacier_everywhere_test.go` walks the SDK's import graph and fails the build if any of the framework's 15 leaf packages stops being used. Drift gets caught before release.

## What is next

- Per-command pages: see <a href="/sdk/commands/">/sdk/commands/</a> for the full reference of each verb.
- Configuration: see <a href="/sdk/configuration">/sdk/configuration</a> for the 6 config keys and every env var.
- Install paths: see <a href="/sdk/install">/sdk/install</a> for `go install`, `go build`, and tagged releases.

The framework libraries are the primary artifact at `github.com/nathanbrophy/glacier`. The SDK is the icing on the cake. Both ship together at v0.

<style scoped>
.sdk-cmd-table {
  width: 100%;
  border-collapse: collapse;
  margin: 1.5rem 0;
}
.sdk-cmd-table th,
.sdk-cmd-table td {
  border: 1px solid var(--vp-c-divider);
  padding: 0.55rem 0.85rem;
  text-align: left;
  vertical-align: top;
}
.sdk-cmd-table th {
  background: var(--vp-c-bg-soft);
  font-weight: 600;
}
.sdk-cmd-table tr td:first-child {
  font-size: 0.85em;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--vp-c-brand-1);
  font-weight: 600;
  white-space: nowrap;
  width: 8rem;
}
.sdk-cmd-table tr td:nth-child(2) {
  font-family: var(--vp-font-family-mono);
  white-space: nowrap;
  font-size: 0.95em;
}
.cast-frame {
  margin: 1.5rem auto;
  max-width: 760px;
  text-align: center;
}
.cast-frame img {
  max-width: 100%;
  border-radius: 8px;
  border: 1px solid var(--vp-c-divider);
  background: #1a1a1d;
}
.cast-frame .cast-caption {
  margin-top: 0.5rem;
  font-size: 0.85em;
  color: var(--vp-c-text-2);
}
</style>
