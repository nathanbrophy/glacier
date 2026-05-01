---
id: 0001
title: Brand Identity
slug: brand-identity
status: accepted
owner-agent: magpie
created: 2026-05-01
last-updated: 2026-05-01
supersedes: []
superseded-by: null
reviewers:
  - { agent: magpie, required: true, signed-off-at: 2026-05-01T00:00:00Z }
  - { agent: otter,  required: true, signed-off-at: 2026-05-01T00:00:00Z }
  - { agent: octopus, required: false, signed-off-at: null }
implementing-commits: []
verified-at: null
docs-extract:
  - public-summary
  - mental-model
  - api
  - examples
  - faq
---

# Brand Identity

## Public Summary

Mongoose is a Go SDK that handles the plumbing so you can focus on what's yours. Like the animal it's named for, mongoose is small, alert, and fearless about the messy parts: argument parsing, configuration layering, lifecycle and signal handling, mock-driven testing, and HTTP transport faking. You write the logic. Mongoose keeps the den safe.

## Mental Model

The mental model is a single sharp line: code unique to your problem stays on your side; everything generic stays on mongoose's side. You never write your own flag parser. You never invent your own config-layering rules. You never hand-roll a mock for an interface you control. When the boundary feels wrong — when something that *feels* generic is on your side, or something that *feels* unique is on mongoose's side — that's a spec to file.

```mermaid
flowchart LR
  subgraph Yours [What's yours]
    A[Handler logic]
    B[Domain types]
    C[Business rules]
  end

  subgraph Mongoose [What mongoose owns]
    D[CLI parsing]
    E[Config layering]
    F[Signal handling]
    G[Mock infrastructure]
    H[HTTP transport]
    I[Sandbox primitives]
  end

  A --> D
  A --> E
  A --> F
  C --> G
  C --> H
```

## Goals

- Define the canonical name story, voice, and tone of mongoose.
- Lock the ASCII logo (banner + wordmark) as committed bytes in `assets/logo/`.
- Lock the color palette as a semantic token system grounded in meaningful color theory.
- Lock the typography pairing for the public site.
- Codify Go and CLI naming conventions referenced by every later spec.
- Codify error-message conventions across the library and CLI registers.
- Articulate the Promise — the four user-facing feel-statements every component spec is tested against.

## Non-Goals

- The public site implementation. That gets its own spec when the site is built.
- Light-mode color tokens. Derived from the dark tokens at site-spec time.
- Localization or translation policy. Deferred to a future spec.
- Branded merchandise, social-media cards, or external assets beyond the repo. Out of scope.
- The CLI module's actual API. That lives in spec 0002 (SDK shape) and the CLI component spec that follows.

## Architecture

Identity is a taxonomy and a surface map, not a code architecture.

**Taxonomy** — five identity layers, each owned by a specific artifact in the repo:

| Layer | Owned by | Surface |
|---|---|---|
| Voice & tone | This spec | All prose: README, site, doc-comments, error messages, help text |
| Visual identity | This spec; art under `assets/logo/` | ASCII logo, palette tokens, typography |
| Naming conventions (Go) | This spec | Every package, type, function, error, interface in the SDK |
| Naming conventions (CLI) | This spec | Every CLI verb, flag, subcommand emitted by the CLI module |
| The Promise | This spec | Acceptance criterion for every component spec |

**Surface map** — where each identity element appears:

```
README.md                  → banner ASCII + tagline + Public Summary (extracted by Magpie)
public site (future)       → all five identity layers
specs/_template.md         → none (template is structural; identity applies inside spec content)
.claude/agents/*.md        → voice & tone (so agents speak in mongoose's register)
mongoose --version         → wordmark + tagline (rendered via the CLI module's banner feature, per dogfooding)
mongoose --help (CLI mod)  → CLI naming + help-text style
package godoc              → Go naming + voice
error.Error() throughout   → library error-message style
CLI error output           → CLI error-message style
```

## Schema

This spec introduces no Go types. The "schema" of the identity is the semantic-token table for color, the typography table, and the naming-convention tables — all under `## Examples` below, since they are also the most useful didactic content.

## API

N/A — this spec introduces no Go API. The CLI module spec (later) will introduce APIs governed by the naming conventions defined here.

## Examples

### Banner ASCII (canonical)

The banner pairs a small ASCII mongoose mascot with the wordmark. Committed verbatim to `assets/logo/banner.txt` and reproduced here:

```text
                          ███╗   ███╗ ██████╗ ███╗   ██╗ ██████╗  ██████╗  ██████╗ ███████╗███████╗
       .-""""-.            ████╗ ████║██╔═══██╗████╗  ██║██╔════╝ ██╔═══██╗██╔═══██╗██╔════╝██╔════╝
      / -    - \           ██╔████╔██║██║   ██║██╔██╗ ██║██║  ███╗██║   ██║██║   ██║███████╗█████╗
     ( =o    o= )          ██║╚██╔╝██║██║   ██║██║╚██╗██║██║   ██║██║   ██║██║   ██║╚════██║██╔══╝
      \    --  /__         ██║ ╚═╝ ██║╚██████╔╝██║ ╚████║╚██████╔╝╚██████╔╝╚██████╔╝███████║███████╗
       \_______/  \        ╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═══╝ ╚═════╝  ╚═════╝  ╚═════╝ ╚══════╝╚══════╝
                                          Less plumbing. More Go.
```

The mascot is pure ASCII (no Unicode), 5 lines × ~16 cols, so it survives dumb terminals and stripped-encoding contexts. The wordmark uses the ANSI Shadow style with Unicode box-drawing characters, 6 lines × ~75 cols, sized to fit the 80-col README convention with margin. The mascot sits visually centered against the wordmark (one line of leading above it). The tagline sits centered beneath the wordmark.

### Compact wordmark (canonical)

The compact wordmark is wordmark + tagline only — no mascot. It's the form rendered by `mongoose --version`. Committed verbatim to `assets/logo/wordmark.txt`:

```text
███╗   ███╗ ██████╗ ███╗   ██╗ ██████╗  ██████╗  ██████╗ ███████╗███████╗
████╗ ████║██╔═══██╗████╗  ██║██╔════╝ ██╔═══██╗██╔═══██╗██╔════╝██╔════╝
██╔████╔██║██║   ██║██╔██╗ ██║██║  ███╗██║   ██║██║   ██║███████╗█████╗
██║╚██╔╝██║██║   ██║██║╚██╗██║██║   ██║██║   ██║██║   ██║╚════██║██╔══╝
██║ ╚═╝ ██║╚██████╔╝██║ ╚████║╚██████╔╝╚██████╔╝╚██████╔╝███████║███████╗
╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═══╝ ╚═════╝  ╚═════╝  ╚═════╝ ╚══════╝╚══════╝
                       Less plumbing. More Go.
```

### Color palette (canonical hex, with role)

The palette is dark-first. Each token is bound to a role; colors are never used decoratively.

| Token | Hex | Role | When used |
|---|---|---|---|
| `--mg-bg` | `#0E1116` | Page background | Default canvas |
| `--mg-surface` | `#161B22` | Elevated surface | Cards, code blocks, panels |
| `--mg-surface-2` | `#1F262E` | Hovered or focused surface | Hover, focus, active state |
| `--mg-text` | `#E6EDF3` | Primary text | Body copy, headings |
| `--mg-text-muted` | `#8B949E` | Secondary text | Captions, metadata, secondary labels |
| `--mg-text-faint` | `#6E7681` | Tertiary text | Disabled, deemphasized |
| `--mg-cyan` | `#22D3EE` | Primary accent | Links, primary CTAs, brand emphasis |
| `--mg-teal` | `#2DD4BF` | Secondary accent | Hover states, secondary CTAs |
| `--mg-success` | `#4ADE80` | Success state | Pass, confirm, ready |
| `--mg-warning` | `#FBBF24` | Caution state | Pending, deprecation, soft alert |
| `--mg-error` | `#F87171` | Failure state | Fail, destructive, error output |
| `--mg-info` | `#7DD3FC` | Neutral info | Notes, hints, tooltips |
| `--mg-border` | `#30363D` | Hairlines | Dividers, table borders |

**Color theory.** The palette is cool-on-cool: cyan and teal accents on a near-black background. To prevent the cold, sterile look of pure cyan-on-black, the primary text token is a *warm* off-white (`#E6EDF3`, with slight pink-yellow lift). State colors (success, warning, error, info) are deliberately desaturated relative to the brand accents so they don't compete when shown alongside. Cyan (`#22D3EE`) and teal (`#2DD4BF`) are split-complementary within the cool spectrum — enough variation to distinguish primary from secondary without fighting each other. Contrast: `--mg-text` on `--mg-bg` is 15.9:1 (AAA, well above 7:1); `--mg-cyan` on `--mg-bg` is 9.6:1 (AAA). Every token meets WCAG AA at minimum.

### Typography pairing (canonical)

| Use | Family | Weights | Source |
|---|---|---|---|
| Display / headings | Space Grotesk | 500, 700 | rsms.me, SIL OFL |
| Body / UI | Inter | 400, 500, 600 | rsms.me, SIL OFL |
| Code | JetBrains Mono | 400, 700 | jetbrains.com/mono, SIL OFL |

All three are vendored when the public site lands — never CDN-loaded. The site spec will commit the actual font files.

### Sample Go following the conventions

```go
// Package cli builds production-ready command-line interfaces from idiomatic handler functions.
package cli

import (
	"context"
	"errors"
)

// ErrCancelled is returned when a command is interrupted by a signal before completion.
var ErrCancelled = errors.New("cli: command cancelled")

// ParseError is returned when an argument or flag fails to parse. It wraps the original
// parser error and identifies the offending argument by name.
type ParseError struct {
	Arg string
	Err error
}

// Error implements error. The format is "cli: parse <arg>: <cause>": lowercase, no
// trailing punctuation, per mongoose library-error conventions.
func (e *ParseError) Error() string { return "cli: parse " + e.Arg + ": " + e.Err.Error() }

// Unwrap supports errors.Is and errors.As traversal.
func (e *ParseError) Unwrap() error { return e.Err }

// Runner runs a parsed command tree to completion. Implementations must respect ctx
// cancellation: when ctx is done, Run returns ctx.Err() wrapped in ErrCancelled.
type Runner interface {
	Run(ctx context.Context) error
}
```

This block demonstrates: package-name idiom (single short word, lowercase, no underscores); no-stutter exported type names; `Err<Cause>` sentinel; `<Cause>Error` typed error; canonical error-message format; `Unwrap` for `errors.Is`/`As`; single-method `Runner` interface with `<Verb>er` naming; `ctx context.Context` first parameter on a cancellable function; doc comments starting with the symbol name.

### Sample CLI invocations following the conventions

```text
$ mongoose serve --port 8080 --enable-metrics --quiet
$ mongoose build --output ./dist
$ mongoose init my-project
$ mongoose --help
```

Verbs are imperative and lowercase. Flags use `--long-kebab-case`. Boolean flags use the `--enable-x` / `--disable-x` form, never `--x=true`. Short flags exist only for the most common operations.

A CLI error follows the user-facing register — capitalized, problem + cause + actionable next step, ends with a period:

```text
Error: Cannot bind to port 8080: address already in use.
       Try a different port with --port, or stop the process holding 8080.
```

The same condition surfaced as a *library* error to a Go caller follows the library register — lowercase, no trailing punctuation, `package: action: cause`:

```go
err := server.Listen(ctx, ":8080")
// err.Error() == "cli: serve: bind tcp :8080: address already in use"
```

Two registers, two audiences, one underlying condition.

### Naming convention reference (Go)

| Element | Convention | Example |
|---|---|---|
| Package name | short, lowercase, single word, no underscores; no plural unless genuinely a collection | `cli`, `mock`, `httpmock`, `sandbox`, `config` |
| Exported type | PascalCase, no stutter (`cli.Command`, never `cli.CLICommand`) | `Command`, `Builder`, `Mock` |
| Function | PascalCase; verb-first for actions; noun-only for accessors (no `Get` prefix) | `Run(ctx)`, `Build()`, `Name()` |
| Sentinel error | `Err<Cause>` | `ErrCancelled`, `ErrInvalidConfig` |
| Typed error | `<Cause>Error`, with `Unwrap` | `ParseError`, `ValidationError` |
| Single-method interface | `<Verb>er` | `Reader`, `Builder`, `Closer` |
| Multi-method interface | descriptive noun | `Command`, `Transport` |
| Doc comment | full sentences, start with the symbol name | `// Command represents a parsed command tree.` |
| Cancellable function | `ctx context.Context` first parameter | `func Run(ctx context.Context) error` |

### Naming convention reference (CLI)

| Element | Convention | Example |
|---|---|---|
| Verb | imperative mood, lowercase, single word preferred | `build`, `run`, `init`, `serve` |
| Subcommand nesting | flat preferred; two-level max | `mongoose mock generate`, not `mongoose tools mock generate new` |
| Long flag | `--long-kebab-case` | `--enable-metrics`, `--output-dir` |
| Short flag | only for the most common operations | `-h` (`--help`), `-v` (`--version`), `-q` (`--quiet`) |
| Boolean flag | `--enable-x` / `--disable-x` | `--enable-metrics`, never `--metrics=true` |
| Help summary | starts capitalized, ends with period, max 80 cols | `Build a project for production.` |
| Error output | capitalized, problem + cause + next step, ends with period | (see CLI error example above) |

### The Promise

When a developer uses mongoose, they should be able to say each of these truthfully:

1. *"I'm only writing what's mine."*
2. *"I trust the defaults."*
3. *"The error tells me what to do next."*
4. *"Tests are easy because the framework helps."*

Every component spec — CLI, mocks, HTTP transport, sandbox, primitives — is reviewed against these four statements. If the design doesn't deliver them, the design is wrong.

## Test Matrix

| Scenario | Input | Expected | Covered by |
|---|---|---|---|
| Banner ASCII rendering | `cat assets/logo/banner.txt` in any modern monospace terminal | Renders identically to the canonical bytes above | Visual review during spec verification |
| Wordmark ASCII rendering | `cat assets/logo/wordmark.txt` in any modern monospace terminal | Renders identically to the canonical bytes above | Visual review during spec verification |
| Mascot ASCII purity | grep mascot region (lines 1–6, cols 1–18 of `banner.txt`) for non-ASCII codepoints | No matches | `LC_ALL=C grep -P '[^\x00-\x7F]'` over the mascot region returns empty |
| Color contrast (AA) | Each text token vs `--mg-bg` and `--mg-surface` | ≥ 4.5:1 for text, ≥ 3:1 for non-text | Manual computation captured in §Examples; re-verified by site spec |
| Go naming convention adherence | Any package introduced by any later spec | Conventions in the Go reference table apply | Reviewer checklist on every component-spec PR |
| CLI naming convention adherence | Any CLI verb or flag introduced by any later spec | Conventions in the CLI reference table apply | Reviewer checklist on the CLI module spec and downstream |
| Error-message register | Every `error.Error()` and CLI error output in the codebase | Library register lowercase no-trailing-period; CLI register capitalized period actionable | Reviewer checklist; example assertions added to CLI module test suite when the module ships |
| Voice / no-superlatives | Any committed prose under `README.md`, `specs/`, `.claude/agents/`, future `site/` | Zero occurrences of "blazing", "revolutionary", "best-in-class", "amazing", "seamless" | grep audit run during this spec's verification; later: optional CI lint |
| Promise satisfaction | Every component spec | The component lets a developer say each of the four Promise statements truthfully | Reviewer checklist (Magpie + Otter on every component spec) |
| Dogfooding | Any new public feature in the CLI module (when the module ships) | The feature is exercised by at least one mongoose-CLI command | Reviewer checklist; Lynx test matrix on the CLI module spec |

## Dependency Justification

Empty.

| Module | Version | License | Last release | Maintainers | Alternatives | Why we can't roll our own |
|---|---|---|---|---|---|---|
|  |  |  |  |  |  |  |

Fonts on the eventual public site are vendored, not Go dependencies; their justification lives in the site spec.

## Security & Supply-Chain Notes

- No untrusted input is handled by the artifacts of this spec.
- The ASCII-art logo files contain no secrets and are safe to commit publicly.
- When the public site is built, the typography vendoring decision (vendored vs. CDN) lands under the site spec's security review. The decision is pre-committed in this spec's Decisions & Rationale: vendored only.

## Migration & Compatibility

N/A. This is the first identity spec; nothing to migrate from.

## FAQ

**Why is the project named "mongoose"?**
Because a mongoose is a small keeper. It's alert, fearless against bigger threats, social, and good at handling pests. That's the shape of the SDK: small surface, fast feedback, fearless about the messy generic problems, and a den-keeper for the handler code that's actually yours.

**Why dark-first?**
Mongoose is a developer tool. Most developers live in dark terminals and dark IDEs. A dark-first identity is native to the environment the SDK runs in. Light mode will be defined by the site spec when the site is built; dark is the source of truth.

**Why these specific naming conventions?**
They're not bespoke — they're Effective Go and the Go standard library, codified in one place so contributors and agents have a single canonical reference. Mongoose has zero novel naming conventions.

**Can I theme the CLI output?**
Yes — when the CLI module spec lands, the palette tokens will be exposed as terminal-color overrides. Default behavior uses the dark-first cyan/teal palette; users can set `MONGOOSE_NO_COLOR=1` to disable, or override individual tokens via env vars.

**Why ban superlatives?**
Because "blazing fast" doesn't tell a developer anything. "Parses 50,000 flags per second on an M2 Air" does. The ban forces specifics. Specifics are what make documentation useful.

**Why two ASCII-logo variants instead of one?**
Because `--version` output is a tight context (one or two terminal screens) and a full banner with a mascot would dominate it. The compact wordmark fits gracefully there. The full banner has room to breathe in the README header and the site hero.

**Does mongoose's own CLI use the mongoose CLI module?**
Yes — it must. Mongoose ships a `mongoose` CLI binary built entirely with its own CLI module. Anywhere the banner is rendered at runtime — `mongoose --version`, welcome screens, error contexts — that rendering goes through the CLI module's banner feature, never via bespoke code. If the CLI module can't render the banner correctly, that's a bug in the CLI module, surfaced immediately by mongoose's own tooling. Dogfooding is the project's strongest credibility signal.

## Decisions & Rationale

This spec was authored from a fully-resolved plan. Each decision is recorded with its rationale.

- **D1 — Name angle: the keeper.** The mongoose-as-keeper framing maps cleanly to the SDK's promise and avoids competitor framing (which is forbidden by the bootstrap).
- **D2 — Tagline: "Less plumbing. More Go."** Two short noun phrases, strong rhythm, says exactly what the SDK does.
- **D3 — Two ASCII logo variants.** Banner for spacious contexts (README, site hero), compact wordmark for tight contexts (`--version`).
- **D4 — Wordmark style: ANSI Shadow.** High visual weight; renders identically in any modern monospace terminal.
- **D5 — Mascot style: minimalist ASCII silhouette.** Pure ASCII (no Unicode) so the mascot survives dumb terminals; 6×16 to align with the wordmark.
- **D6 — Palette: dark-first cyan/teal accents, encoded as semantic tokens.** Each token bound to a role, never used decoratively. Light-mode tokens deferred to the site spec.
- **D7 — Typography: Inter, Space Grotesk, JetBrains Mono.** Open-source, performant, dev-tool-native, complementary personalities.
- **D8 — Go naming conventions.** Effective Go and the standard library, codified here as the single canonical reference.
- **D9 — CLI naming conventions.** Patterns developers expect from `git`, `kubectl`, `cargo`. POSIX/GNU long-option compatible.
- **D10 — Error-message conventions, two registers.** Library errors and CLI errors serve different audiences; consistent style within each.
- **D11 — Voice & tone.** Direct, second person, examples-first, honest about tradeoffs, no superlatives. The ban on superlatives forces specifics.
- **D12 — The Promise.** Four developer-feel statements that every component spec is reviewed against. The user-facing translation of the mission.
- **D13 — Logo art committed as plaintext under `assets/logo/`.** Single source of truth; no regeneration at build time. Future CLI module embeds via `//go:embed`.
- **D14 — Identity adds zero Go dependencies.** Fonts are vendored when the site is built, not loaded from a third-party CDN.
- **D15 — Dogfooding is a project commitment.** Mongoose ships a `mongoose` CLI binary built with its own CLI module. The brand-identity assets are consumed by that binary via the CLI module's banner feature; bespoke rendering is forbidden.

## Open Questions

None. Every question raised during the design conversation is resolved in Decisions & Rationale above.

## Verification

Run, in order:

1. Open `assets/logo/banner.txt` and `assets/logo/wordmark.txt` in a modern monospace terminal (Windows Terminal, iTerm2, Alacritty, gnome-terminal, etc.). Confirm rendering matches the bytes in `## Examples` byte-for-byte.
2. Run `LC_ALL=C grep -P '[^\x00-\x7F]'` over the mascot region (lines 1–6, columns 1–18) of `assets/logo/banner.txt`. Expect zero matches: the mascot is pure ASCII.
3. Confirm WCAG contrast: `--mg-text` on `--mg-bg` ≥ 7:1 (AAA), `--mg-cyan` on `--mg-bg` ≥ 4.5:1 (AAA), every other text token on `--mg-bg` ≥ 4.5:1 (AA).
4. Confirm `README.md` opens with the banner ASCII, followed by the tagline, followed by the Public Summary.
5. Search the repo for forbidden superlatives:
   ```sh
   grep -rEi 'blazing|revolutionary|best-in-class|amazing|seamless' --include='*.md' --include='*.txt' .
   ```
   Expect zero matches in committed prose.
6. Confirm Magpie's and Otter's `signed-off-at` fields in the front matter are non-null and `## Open Questions` is empty.

If any check fails, the spec returns to `in-review` and the issue is fixed before re-acceptance.
