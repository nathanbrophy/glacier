---
status: research
owner: magpie
created: 2026-05-03
---

# SDK Doc Surface

Research artifact supplying voice content, phrase rotations, error strings, vibe tips,
mascot briefs, README prose, and site page drafts for spec 0032-sdk. Every concrete
identifier, marker, and API name is grounded in spec 0032, spec 0011, spec 0016, or
spec 0001 text. Nothing is invented.

---

## 1. Verb-and-Flag Review

### Methodology

Each command verb and each flag is tested against spec 0001 CLI naming conventions:
- Verbs: imperative mood, lowercase, single word preferred.
- Long flags: `--long-kebab-case`.
- Boolean flags: `--enable-x` / `--disable-x` pattern (or positive-form `--no-x` for opt-out).
- Help summaries: capitalized, ends with period, max 80 cols.

### Command verbs

| Verb | Spec 0001 compliant? | Notes |
|---|---|---|
| `vibe` | Yes | Imperative, lowercase, single word. |
| `version` | Yes | Noun used as verb in the "print version" idiom; canonical (git, cargo, go). |
| `generate` | Yes | Imperative, lowercase, single word. |
| `lint` | Yes | Imperative, lowercase, single word. |
| `test` | Yes | Imperative, lowercase, single word. |
| `init` | Yes | Imperative, lowercase, single word. |
| `new` | Yes | Imperative, lowercase, single word; parent of `package`, `command`, `option`. |
| `completions` | Propose rename to `completion` | "completions" is a plural noun. Spec 0001 prefers "imperative mood, single word preferred." The canonical git, cargo, and kubectl usage is `completion` (singular). However, the spec locks the verb as `completions` in D-S8 and the `CompletionsCmd` struct. Per the "do not second-guess locked decisions" instruction, I record the observation but do not override the lock. **No change proposed.** |
| `explain` | Yes | Imperative, lowercase, single word. |

All nine verbs comply with spec 0001. No rename proposed.

### Subcommand verbs under `new`

| Verb | Spec 0001 compliant? | Notes |
|---|---|---|
| `package` | Yes | Single word, lowercase. |
| `command` | Yes | Single word, lowercase. |
| `option` | Yes | Single word, lowercase. |

### Flags

#### Global / auto-injected flags (D-S33)

| Flag | Compliant? | Notes |
|---|---|---|
| `--help` / `-h` | Yes | Canonical. |
| `--version` / `-v` | Yes | Canonical; spec 0001 permits `-v` for `--version`. |
| `--quiet` / `-q` | Yes | Canonical. |
| `--verbose` / `-V` | Yes | Single word, kebab not needed. |
| `--very-verbose` | Yes | Kebab-case compound. |
| `--no-animate` | Yes | `--no-x` boolean opt-out form; compliant. |
| `--no-banner` | Yes | Same pattern. |
| `--profile` | Yes | Single word; value-taking, not boolean. |
| `--otel-endpoint` | Yes | Kebab-case compound. |

#### `vibe`

| Flag | Compliant? | Notes |
|---|---|---|
| `--duration` | Yes | Single word. |
| `--no-tips` | Yes | `--no-x` opt-out form. |
| `--seed` | Yes | Single word. |
| `--ascii` | Yes | Single word. |

#### `version`

| Flag | Compliant? | Notes |
|---|---|---|
| `--check` | Yes | Single word. |
| `--strict` | Yes | Single word. |
| `--json` | Yes | Single word. |

#### `generate`

| Flag | Compliant? | Notes |
|---|---|---|
| `--check` | Yes | Consistent across commands that support drift detection. |
| `--only` | Yes | Single word. |
| `--parallel` | Yes | Single word. |
| `--no-status` | Yes | `--no-x` opt-out form. |

Note: `Patterns` is a positional argument, not a flag. Compliant.

#### `lint`

| Flag | Compliant? | Notes |
|---|---|---|
| `--fix` | Yes | Single word. |
| `--severity` | Yes | Single word. |
| `--format` | Yes | Single word. |
| `--no-cache` | Yes | `--no-x` opt-out form. |

#### `test`

| Flag | Compliant? | Notes |
|---|---|---|
| `--race` | Yes | Single word. |
| `--cover` | Yes | Single word. |
| `--fuzz` | Yes | Single word. |
| `--bench` | Yes | Single word. |
| `--baseline` | Yes | Single word. |
| `--update-baseline` | Yes | Kebab-case compound. |
| `--format` | Yes | Consistent with `lint`. |
| `--slowest` | Yes | Single word. |
| `--no-status` | Yes | Consistent with `generate`. |

#### `init`

| Flag | Compliant? | Notes |
|---|---|---|
| `--name` | Yes | Single word. |
| `--template` | Yes | Single word. |
| `--license` | Yes | Single word. |
| `--mascot` | Yes | Single word. |
| `--no-git` | Yes | `--no-x` opt-out form. |
| `--yes` / `-y` | Yes | Single word; short form documented. |
| `--force` | Yes | Single word. |

Note: `Dir` is a positional argument. Compliant.

#### `new`

| Flag | Compliant? | Notes |
|---|---|---|
| `--dry-run` | Yes | Kebab-case compound; canonical idiom. |
| `--force` | Yes | Single word; consistent. |
| `--parent` | Yes | Single word (on `new command`). |
| `--package` | Yes | Single word (on `new option`). |

#### `completions`

| Flag | Notes |
|---|---|
| `Shell` (positional) | No flags; shell is positional argument. Compliant. |

#### `explain`

| Flag | Notes |
|---|---|
| `Topic` (positional) | Positional. |
| `--list` | Yes. Single word boolean. |

### Summary

No renames proposed for verbs or flags. All comply with spec 0001.

The one observation worth recording for future spec consideration: `completions` (plural)
deviates from the single-word imperative norm that applies to all eight other commands.
The lock in D-S8 stands; this note is for the record only.

---

## 2. Success-Tone Rotation Table

Per D-S17: lightly playful rotation, max one phrase per invocation, paired with confident
kaomoji `ʕ⌐■-■ʔ`, max 6 words, calm + lightly playful per spec 0001 D11.

Starting points given in spec: "nice", "all set", "ready".

| # | Phrase | Word count | Notes |
|---|---|---|---|
| 1 | nice. | 1 | Spec-canonical; minimal, warm. |
| 2 | all set. | 2 | Spec-canonical; calm reassurance. |
| 3 | ready. | 1 | Spec-canonical; confident one-word. |
| 4 | done and dusted. | 3 | Colloquial, light. |
| 5 | looking good from here. | 4 | Companion-like; present tense. |
| 6 | clean run. | 2 | Terse, positive, technical register. |
| 7 | that went well. | 3 | Understated; honest. |
| 8 | on solid ground. | 3 | Glacier metaphor; grounded. |
| 9 | good to go. | 3 | Idiomatic; common developer phrase. |
| 10 | nothing to complain about. | 4 | Wry, honest; anti-superlative. |

Each phrase is lowercase, ends with a period, is <=6 words, and avoids superlatives.
The rotation is seeded deterministically per the `--seed` flag (D-S56).

Usage context per phrase:
- `nice.` and `clean run.` suit `generate` and `test` (completion of a mechanical task).
- `all set.` and `ready.` suit `init` (project now ready to develop).
- `done and dusted.` suits `lint` (no findings).
- `looking good from here.` suits `version --check` when current = latest.
- `that went well.` suits `test` with full green run.
- `on solid ground.` suits `init` or `new` after scaffold writes.
- `good to go.` suits `generate` with no drift.
- `nothing to complain about.` suits `lint` with zero findings.

---

## 3. Error String Structures Per Command

Per D-S18 and D-S74: calm + actionable. Structure: state the failure, show the cause,
suggest the next step, offer `glacier explain <code>` link.
Per spec 0001 CLI error register: capitalized, problem + cause + actionable next step,
ends with a period.

Format used below: `Error: <Problem>: <cause>. <Next step>. Run \`glacier explain <N>\` for details.`

Exit codes are from D-S27.

### vibe (most likely failure: raw-mode acquire fails)

```
Error: Cannot enter raw mode: <cause>.
       Try `--ascii` for the static fallback, or check terminal permissions.
       Run `glacier explain 70` for details.
```

Exit 70.

### version (most likely failure: offline with --strict)

```
Error: Cannot reach GitHub Releases for nathanbrophy/glacier: <cause>.
       Check your network connection and try again, or omit --strict to degrade gracefully.
       Run `glacier explain 68` for details.
```

Exit 68.

### generate (most likely failure: generator returns error)

```
Error: cli/gen failed: <cause>.
       Check the marker syntax in the affected file, then re-run `glacier generate`.
       Run `glacier explain 64` for details.
```

Exit 64.

### generate --check (drift detected)

```
Error: 1 generated file is stale.
       Run `glacier generate` to update, then commit the result.
       Run `glacier explain 69` for details.
```

Exit 69.

### lint (most likely failure: findings at or above severity threshold)

```
Error: 2 errors, 3 warnings in ./...: see findings above.
       Fix errors to proceed; use --severity=warning to report lower-severity issues.
       Run `glacier explain 65` for details.
```

Exit 65.

### test (most likely failure: tests fail)

```
Error: conf.TestLayerConflict_File_vs_Env failed.
       Run the failing test in isolation: glacier test ./conf/ -run TestLayerConflict_File_vs_Env -v
       Run `glacier explain 66` for details.
```

Exit 66.

### init (most likely failure: existing files without --force)

```
Error: Cannot write to my-app/: 3 file(s) exist.
       Use --force to overwrite existing files after reviewing the conflict list above.
       Run `glacier explain 67` for details.
```

Exit 67.

### new (most likely failure: no go.mod found)

```
Error: No Go module found in my-app/ or any ancestor directory.
       Run `glacier init` first to scaffold a module, then re-run `glacier new`.
       Run `glacier explain 67` for details.
```

Exit 67.

### completions (most likely failure: unknown shell)

```
Error: Unknown shell "ksh": supported shells are bash, zsh, fish, pwsh.
       Run `glacier explain 2` for usage-error details.
```

Exit 2.

### explain (most likely failure: unknown topic)

```
Error: Unknown topic "exit-999". Did you mean "exit-66"?
       Run `glacier explain --list` to see all available topics.
```

Exit 2.

---

## 4. The 12 Vibe Tips

Per D-S59: 12 tips, categories locked (option, cli, term, log, conf, errs, assert, fixture,
mock, httpmock, httpc, concur), body 16-200 printable ASCII, spec citation included.
Otter reviews technical accuracy after this draft.

The placeholder rows from spec 0032 are rewritten here in Glacier voice: direct, second
person, one verifiable claim, calm tone, no superlatives.

| # | Category | Body | Spec ref |
|---|---|---|---|
| 1 | option | option.Apply is goroutine-safe and produces zero allocations when the option slice is empty. | spec 0003 §API |
| 2 | cli | glaciergen never links cli/gen into your binary. Build it, use it, remove it: the binary still runs. | spec 0011 §Mental Model |
| 3 | term | term.Animator intercepts every slog record while an animation runs so logs and progress bars never collide on screen. | spec 0016 §Public Summary |
| 4 | log | Log records on a TTY show color-only level prefixes. The kaomoji is reserved for command-boundary status lines only. | spec 0032 §Architecture D-S15 |
| 5 | conf | conf.Load is atomic: concurrent readers always see a complete, consistent struct, never a partial write. | spec 0009 §Mental Model |
| 6 | errs | errs.Sentinel panics at startup if your message does not match the library register format. The check is a unit test. | spec 0004 §API |
| 7 | assert | assert.Equal[T] is order-insensitive for maps and dereferences pointers before comparing, with no reflect.DeepEqual overhead on the fast path. | spec 0006 §API |
| 8 | fixture | fixture.Golden lets you bless expected output once: set GLACIER_GOLDEN_UPDATE=1 and run go test, then commit. | spec 0010 §API |
| 9 | mock | mock.Of[T] is typed at compile time. Runtime reflection is used only during test setup, never in the hot path. | spec 0012 §Mental Model |
| 10 | httpmock | httpmock and httpc share no imports. Wire them together at the test boundary; neither knows about the other. | spec 0013 §Mental Model |
| 11 | httpc | httpc body closures are called once per attempt. Retries are correct by construction: no double-read of a closed body. | spec 0015 §Mental Model |
| 12 | concur | concur.Group's WaitDone returns errs.Join over every goroutine error. No goroutine failure is silently swallowed. | spec 0007 §API |

Voice notes per tip:
- All bodies are in second person ("your") or describe the package's behavior as a fact.
- No body uses "blazing", "revolutionary", "amazing", "seamless", or "best-in-class".
- No body uses em-dash.
- Every body is between 16 and 200 printable ASCII characters (verified by inspection).
- The `log` tip cites spec 0032 directly because the behavior (D-S15) is SDK-specific;
  Otter should confirm a spec 0032 citation is appropriate or redirect to the log spec.

---

## 5. Mascot Library Briefs

Per D-S38 and spec 0001 Amendment B: six mascots. The polar bear is Glacier's canonical
mascot (spec 0001 D44). The other five are scoped to user projects scaffolded by
`glacier init`. Each brief includes the kaomoji form, a brand brief paragraph, and a
confirmation that the 5-line block-character form fits <=16 columns.

Kaomoji forms are taken verbatim from spec 0032 §Schema (MascotID constants) and the
`## Documentation Surface` mascot table.

---

### polar_bear

Kaomoji: `ʕ•ᴥ•ʔ`

"Glacier the Bear" is the Glacier framework's canonical mascot per spec 0001 D44. When
you scaffold a project with `glacier init` and pick the polar bear, you are carrying that
same stable, deep, and predictable energy into your own project. The polar bear's wide,
friendly block-character form (5 lines, drawn with `▟ ▀ ▙ █ ● ▼`) appears in the Glacier
banner and in your project's own generated banner once figgen renders your project name
alongside it. Choose the polar bear when you want your project to feel grounded and
unhurried: a keeper of its domain, solid and long-running.

5-line block-character form: 5 lines x approximately 14 columns. Fits within 16 columns.
Confirmed compliant with spec 0001 D44 ("5 lines x ~14 cols").

---

### penguin

Kaomoji: `<(•^•)>`

The penguin is a cold-weather peer of the polar bear: social, upright, and a little
formal. Pick the penguin for a project that values protocol and structure, a server or
a data-pipeline tool where reliability is the main story. The penguin's posture (wings
slightly out, head held level) reads as composed and ready. Its kaomoji uses ASCII
characters only, so it renders cleanly even in terminals that struggle with extended
Unicode ranges, making it the practical choice for teams that run across a wide variety
of CI environments.

5-line block-character form: 5 lines x <=14 columns. Fits within 16 columns.

Note: spec 0032 lists the kaomoji as `<(•^•)>`. The storyboard in §Commands.init
shows `<(•^•)>` as well. Used verbatim here.

---

### owl

Kaomoji: `(o,o)`

The owl is wisdom-coded and minimal. Its block-character form is the most spare of the
six: two wide eyes, a compact body, a suggestion of wings. Pick the owl for a tooling
project where the interface is deliberately quiet and the output is precise, a static
analysis tool, a schema validator, a documentation generator. The owl does not perform;
it observes and reports. Its kaomoji form is short enough to fit inside tight status-line
widths, and its calm expression (no open mouth, no wide paws) matches a tool that speaks
only when it has something worth saying.

5-line block-character form: 5 lines x <=12 columns. Fits within 16 columns.

Note: spec 0032 lists the kaomoji as `(o,o)`. Used verbatim here.

---

### fox

Kaomoji: `^..^`

The fox is curious, fast, and resourceful. Its alternate-ears render distinguishes it
immediately from the bear family. Pick the fox for a CLI tool with a wide command
surface, something that covers many sub-domains, or a project that moves quickly and
ships often. The fox's kaomoji uses simple ASCII carets and dots: it renders everywhere
and communicates personality with minimum characters. A fox project feels like it has
done its homework and is ready to run.

5-line block-character form: 5 lines x <=14 columns. Fits within 16 columns.

Note: spec 0032 lists the kaomoji as `^..^`. Used verbatim here.

---

### otter

Kaomoji: `ʕ•˦•ʔ`

The otter is a stream-dweller: playful, efficient, and at home in flow. Its rounded form
and short ears give it a softer silhouette than the bear. Pick the otter for a project
that is built around data pipelines, streaming APIs, or event-driven workflows, anything
where continuity and smooth throughput are the main qualities. The otter also carries a
nod to Glacier's own Otter agent (the architecture reviewer), which makes it the natural
choice for a project that is itself a framework or a developer-facing library.

5-line block-character form: 5 lines x <=14 columns. Fits within 16 columns.

Note: spec 0032 lists the kaomoji as `ʕ•˦•ʔ`. Used verbatim here.

---

### raccoon

Kaomoji: `(^-ω-^)`

The raccoon is a forest peer with a banded mask and a reputation for problem-solving.
Pick the raccoon for a project that works in the margins: a utility belt, a migration
tool, a script runner, anything that handles the awkward cases and adapts to whatever
environment it finds itself in. The raccoon's mask gives it a distinct, recognizable
shape in block-character form, and its kaomoji reads as gently amused, a good tone for
a tool that knows the domain it is working in is messy and does not pretend otherwise.

5-line block-character form: 5 lines x <=14 columns. Fits within 16 columns.

Note: spec 0032 lists the kaomoji as `(^-ω-^)`. Used verbatim here.

---

## 6. README Narrative for the SDK Section

Per spec 0032 §Documentation Surface §README narrative refresh. Tone: calm + companion-like,
no superlatives, direct second person. The spec provides a skeleton; this is the
voice-polished version.

---

```markdown
## Glacier SDK

The Glacier SDK is `glacier`, a CLI binary built on every Glacier framework package.
It covers the developer day in nine commands and serves as the framework's longest-running
integration test: every package Glacier ships is exercised by at least one SDK command.
A CI gate fails if any package falls out of use.

### Install

```sh
go install github.com/nathanbrophy/glacier/cmd/glacier@latest
```

### What ships

Nine commands in four groups:

**CREATE**

| Command | What it does |
|---|---|
| `glacier init` | Scaffold a new Glacier project with signal handling, banner, and all wiring in place. |
| `glacier new` | Add a package, command, or functional option to an existing Glacier project. |

**DEVELOP**

| Command | What it does |
|---|---|
| `glacier generate` | Run all Glacier code generators (cli, mock, httpmock) over the current module. |
| `glacier lint` | Run gofmt, go vet, staticcheck, and Glacier-specific lints. |
| `glacier test` | Run the Go test suite with a live status panel and an aggregated summary. |

**INSPECT**

| Command | What it does |
|---|---|
| `glacier vibe` | Play the Glacier vibes animation: dancing polar bear, scrolling wordmark, rotating tips. |
| `glacier version` | Print the current version. Pass `--check` to compare against the latest release. |
| `glacier explain` | Print an explanation for a marker, exit code, or config key. |

**UTILITY**

| Command | What it does |
|---|---|
| `glacier completions` | Print a shell-completion script for bash, zsh, fish, or pwsh. |

### Why dogfood

The SDK is built with the same packages you use in your own projects. When the CLI
builder, the mock generator, or the terminal output layer has a rough edge, the SDK
surfaces it immediately because the SDK lives with that rough edge too. It is the
framework's best-effort proof that the packages work together in production.

Read the full SDK reference at <https://nathanbrophy.github.io/glacier/sdk/>.
```

---

## 7. `/sdk/index.md` Draft

This replaces `site/sdk.md`. Route: `/sdk/`.

---

```markdown
---
title: Glacier SDK
description: A CLI binary that covers the Glacier developer day in nine commands.
---

# Glacier SDK

```
ʕ•ᴥ•ʔ glacier v0.1.0
  build: 2026-05-02T14:22:00Z
  go:    go1.26.0
  os:    darwin/arm64
ʕ⌐■-■ʔ latest: v0.1.2 (released 2026-05-08)
  upgrade: go install github.com/nathanbrophy/glacier/cmd/glacier@latest
```

The Glacier SDK is `glacier`, a CLI binary built on every Glacier framework package.
Install it once and you have a scaffold tool, a code generator, a lint suite, a test
runner, a version checker, a shell-completion writer, an inline reference, and a vibe
animation all in one binary built from the same packages you use in your own projects.

## Install

```sh
go install github.com/nathanbrophy/glacier/cmd/glacier@latest
```

Binaries are published for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, and
windows/amd64. The `go install` path above builds from source; no pre-built binary download
is required.

## Nine commands

The SDK covers the developer day in four groups. Run `glacier --help` to see them all,
or click any command name below for the full reference.

### CREATE

| Command | One-liner |
|---|---|
| [`glacier init`](./commands/init.md) | Scaffold a new Glacier project with all wiring in place. |
| [`glacier new`](./commands/new.md) | Add a package, command, or functional option to an existing project. |

### DEVELOP

| Command | One-liner |
|---|---|
| [`glacier generate`](./commands/generate.md) | Run all Glacier code generators over the current module. |
| [`glacier lint`](./commands/lint.md) | gofmt + go vet + staticcheck + six Glacier-specific lints. |
| [`glacier test`](./commands/test.md) | Live status panel and aggregated summary over `go test -json`. |

### INSPECT

| Command | One-liner |
|---|---|
| [`glacier vibe`](./commands/vibe.md) | The Glacier vibes animation: dancing bear, scrolling wordmark, rotating tips. |
| [`glacier version`](./commands/version.md) | Print the current version. `--check` compares against the latest release. |
| [`glacier explain`](./commands/explain.md) | Explain a marker, exit code, or config key inline. |

### UTILITY

| Command | One-liner |
|---|---|
| [`glacier completions`](./commands/completions.md) | Print a shell-completion script for bash, zsh, fish, or pwsh. |

## Quickstart

```sh
# Install
go install github.com/nathanbrophy/glacier/cmd/glacier@latest

# Scaffold a project
glacier init my-app --yes
cd my-app
go mod tidy

# Run it
go run ./cmd/my-app serve --port 8080
```

Your new `cmd/my-app/main.go` is six lines:

```go
package main

import "github.com/nathanbrophy/glacier/cli"

func main() {
    cli.Default.Main()
}
```

All signal handling, the banner, the `version` and `completions` subcommands, OTEL
initialization, and `httpc` tracing live in the glaciergen-emitted
`cmd/my-app/zz_generated_cli.go`. You do not write any of that. You write your handler.

## Add a command

```sh
glacier new command pause --dry-run   # preview the plan
glacier new command pause             # apply (also re-runs codegen)
go run ./cmd/my-app pause --force
```

## Run the suite

```sh
glacier lint ./...                        # find style and correctness issues
glacier test ./...                        # live panel + summary
glacier test --bench=. --update-baseline  # set the bench baseline
glacier generate --check                  # CI drift gate
```

## Opt into telemetry

The SDK never phones home. When you set `OTEL_EXPORTER_OTLP_ENDPOINT` to a collector
you control, the SDK emits per-command spans (`glacier.cmd.<verb>`) and counters
(`glacier.cmd.runs`, `glacier.cmd.duration`, `glacier.cmd.exit_code`) to that endpoint.
Without the env var, the obs package is a no-op and contributes zero overhead.

```sh
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
glacier test ./...
```

## Why dogfood

The SDK is the framework's longest-running integration test. A CI gate
(`TestGlacierEverywhere`) asserts that every Glacier framework package is imported by at
least one file under `cmd/glacier/...`. If a package drifts out of use, the gate fails.
The SDK and the framework move in lockstep.

When a rough edge exists in the CLI builder, the mock generator, or the terminal output
layer, the SDK surfaces it because the SDK lives with that rough edge too.

## Configuration

The SDK reads a config file at `<UserConfigDir>/glacier/config.json`. All fields have
sensible defaults; the file is optional. Env vars with the `GLACIER__` prefix override
file values; flags override env vars.

Key config entries:

| Key | Default | Effect |
|---|---|---|
| `versioncheck.enabled` | `true` | Enable or disable the `--check` network call. |
| `versioncheck.cache_ttl` | `24h` | TTL for the GitHub Releases cache. |
| `banner.show_on_help` | `true` | Show the banner on `glacier --help`. |
| `palette.override` | `{}` | Override individual color tokens by name. |

Full configuration reference: [/sdk/configuration/](./configuration.md).

## Exit codes

Every SDK command uses a stable exit-code table (D-S27). Use `glacier explain <code>` to
read the meaning of any code inline. Key codes:

| Code | Meaning |
|---|---|
| 0 | Success |
| 2 | Usage error |
| 64 | Generate failed |
| 65 | Lint findings at or above severity threshold |
| 66 | Tests failed or benchmark regressed |
| 67 | Scaffolding failed |
| 68 | Version check unreachable with `--strict` |
| 69 | Codegen drift detected |
| 130 | Interrupted (SIGINT) |
| 143 | Terminated (SIGTERM) |

## Global flags

Every command inherits a set of auto-injected flags from glaciergen:

| Flag | Short | Effect |
|---|---|---|
| `--help` | `-h` | Print help. |
| `--version` | `-v` | Print version. |
| `--quiet` | `-q` | Suppress non-error output and animations. |
| `--verbose` | `-V` | Raise log level to debug. |
| `--very-verbose` | | Raise log level to trace. |
| `--no-animate` | | Force plain output even on a TTY. |
| `--no-banner` | | Suppress the banner on this invocation. |
| `--profile` | | Write CPU, heap, and goroutine pprof files. |
| `--otel-endpoint` | | Override `OTEL_EXPORTER_OTLP_ENDPOINT`. |

## Source

The SDK lives at `cmd/glacier/` in the Glacier repository. Every command is a
`+glacier:command`-annotated struct; the wiring is entirely generated by glaciergen. The
SDK's source is the canonical worked example for building a CLI with the `cli` package.

[View the full command reference](./commands/)
```

---

## 8. `/sdk/commands/index.md` Draft

Route: `/sdk/commands/`. Uses the same CREATE / DEVELOP / INSPECT / UTILITY grouping
from `glacier --help` (D-S39).

---

```markdown
---
title: SDK Commands
description: All nine Glacier SDK commands, grouped by phase of the developer day.
---

# SDK Commands

The Glacier SDK ships nine commands. They map onto four phases of the Go developer day:
CREATE, DEVELOP, INSPECT, and UTILITY. The grouping matches the output of `glacier --help`.

Run any command with `-h` for its short help, or `--help` for the full reference including
the mental model, environment variables, and exit codes.

## CREATE

Commands that bring a project or construct into existence.

### [glacier init](./init.md)

Scaffold a new Glacier project in the current directory (or a named subdirectory).
Walks an interactive prompt sequence (module path, template, license, mascot, git init)
or accepts `--yes` to take all defaults. The generated project's `main.go` is six lines;
all wiring comes from glaciergen.

```sh
glacier init my-app
glacier init my-app --yes
glacier init my-app --template=library-only --license=MIT --no-git
```

### [glacier new](./new.md)

Add a package, a CLI subcommand, or a functional option to an existing Glacier project.
`new command` re-runs codegen after writing so the tree is in a working state immediately.
`--dry-run` shows the plan with unified diffs before committing any write.

```sh
glacier new command pause --dry-run
glacier new command pause
glacier new package store
glacier new option WithTimeout --package=server
```

## DEVELOP

Commands that run during active development and CI.

### [glacier generate](./generate.md)

Run all three Glacier code generators (cli, mock, httpmock) over the current module.
Generators run in parallel inside a `concur.Group` with one status-bar row each.
`--check` mode detects drift and exits 69 without writing any file: the canonical CI gate.

```sh
glacier generate
glacier generate --check
glacier generate --only=cli
glacier generate ./internal/...
```

### [glacier lint](./lint.md)

Run gofmt, go vet, staticcheck (if on PATH), and six Glacier-specific lints over the
current module. Findings are grouped by severity with OSC-8 hyperlinks on capable
terminals. `--fix` applies auto-fixable corrections in place.

```sh
glacier lint ./...
glacier lint --fix ./...
glacier lint --format=sarif ./... > results.sarif
```

### [glacier test](./test.md)

Wrap `go test -json` with a live status panel (active packages with spinner glyphs)
and a boxed summary block at completion (pass/fail counts, coverage vs threshold, slowest
tests, failure details). Bench mode compares against a committed baseline.

```sh
glacier test ./...
glacier test --race --cover ./...
glacier test --bench=. --update-baseline
glacier test --format=junit ./... > results.xml
```

## INSPECT

Commands that give you information about the binary, the framework, or your project state.

### [glacier vibe](./vibe.md)

Run the Glacier vibes animation: dancing polar bear block-character mascot on the left,
ANSI Shadow GLACIER wordmark with a slow gradient shimmer on the right, rotating tip
line below. Tips cycle every 5 seconds from the 12-tip registry. Degrades gracefully
to a one-shot static frame on non-TTY or with `--ascii`.

```sh
glacier vibe
glacier vibe --duration=30s --seed=42
glacier vibe --ascii
```

### [glacier version](./version.md)

Print version, build time, Go toolchain version, and OS/arch. `--check` adds a GitHub
Releases lookup (cached 24h) and prints the latest tag with an upgrade line. Offline
behavior is graceful by default; `--strict` opts into exit 68 on network failure.

```sh
glacier version
glacier version --check
glacier version --check --json | jq .latest
glacier version --check --strict  # for CI
```

### [glacier explain](./explain.md)

Print an explanation for a marker, exit code, or config key. Reads from an embedded
topic registry generated at build time from the spec. The box renders with a title,
body, and "see also" links. `--list` enumerates all available topics.

```sh
glacier explain 66
glacier explain +glacier:command
glacier explain versioncheck.cache_ttl
glacier explain --list
```

## UTILITY

Shell-plumbing commands that integrate the SDK into your terminal environment.

### [glacier completions](./completions.md)

Print a shell-completion script for the named shell. Redirect to the appropriate
shell-completion location; the page for each shell documents the exact path and
reload command.

```sh
glacier completions bash   >> ~/.bash_completion
glacier completions zsh    >> ~/.zsh/completions/_glacier
glacier completions fish   > ~/.config/fish/completions/glacier.fish
glacier completions pwsh   >> $PROFILE
```

## All commands at a glance

| Command | Group | One-liner |
|---|---|---|
| `glacier init` | CREATE | Scaffold a new project. |
| `glacier new` | CREATE | Add a package, command, or option to an existing project. |
| `glacier generate` | DEVELOP | Run all code generators. |
| `glacier lint` | DEVELOP | gofmt + vet + staticcheck + Glacier lints. |
| `glacier test` | DEVELOP | Live test panel and aggregated summary. |
| `glacier vibe` | INSPECT | Glacier vibes animation. |
| `glacier version` | INSPECT | Print version; optionally check for updates. |
| `glacier explain` | INSPECT | Explain a marker, exit code, or config key. |
| `glacier completions` | UTILITY | Shell completion scripts. |

[Back to SDK overview](../index.md)
```

---

## 9. Voice Audit for spec 0032-sdk.md

### Superlatives scan

Search terms: "blazing", "revolutionary", "best-in-class", "amazing", "seamless".

Result: **clean**. No matches in any section of the spec.

### Em-dash scan (U+2014)

Result: **clean**. No em-dash characters found. The spec uses hyphens, colons, and
parenthetical constructions throughout.

### Competitor framing scan

No language frames Glacier against a named competitor. Decisions that reject alternatives
(e.g. D-S26 rejecting auto-update) state the Glacier-side reason, not a competitor comparison.

**Result: clean.**

### Struct-tag usage in spec body (code samples)

The spec 0032 `### new` section contains command struct definitions using struct tags rather
than `// +glacier:` comment markers in two places:

Line range ~1212-1237, `NewPackageCmd` and sibling structs:

```go
type NewPackageCmd struct {
    Name string `glacier:"positional"`
    DryRun bool `glacier:"flag,name=dry-run,default=false"`
    Force  bool `glacier:"flag,default=false"`
    ...
}
```

Per `project_codegen_comments_not_struct_tags.md` memory: this form (`glacier:"flag,..."`)
is the wrong pattern. The canonical pattern is `// +glacier:` comment markers. These struct
definitions are binary-internal (under `cmd/glacier/commands/`) and the spec section is
in `## Commands` which is a PUBLIC-extractable section.

Because these structs appear in a `## Commands` block that Magpie extracts for per-command
pages, shipping them verbatim to `/sdk/commands/new/` would document the wrong syntax.

This is a spec defect, not a voice issue. The fix belongs upstream in spec 0032. I am
recording it here and will not transclude the `new` struct examples verbatim until Otter
updates them to use `// +glacier:` markers. I will open a request to Otter to amend
those struct definitions in spec 0032 before the `new` command page is generated.

### Duplicate section header

The spec body contains a duplicate `### explain` header (one at the end of the `explain`
command sub-section, one immediately following at approximately line 1437). The second
occurrence is empty. This is a spec editing artifact and does not affect voice or naming.
Recording for Otter's attention.

### One potential voice tighten: "longest-running integration test"

The phrase "longest-running integration test" appears in `## Public Summary` and
`## Mental Model` and is repeated in the FAQ and in the README skeleton. The repetition
is deliberate (it is the framing hook for dogfooding) and the phrasing is accurate and
modest, not a superlative. No change required; noting for awareness.

### Summary

| Check | Result |
|---|---|
| Superlatives | Clean |
| Em-dashes | Clean |
| Competitor framing | Clean |
| Struct-tag usage in code samples | Defect found in `### new` structs. Spec amendment needed before extraction. |
| Duplicate section header | Editing artifact in `### explain`. Spec cleanup needed. |
