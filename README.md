<!--
  Banner ASCII is the canonical art from specs/0001-brand-identity.md.
  Source of truth: assets/logo/banner.txt. Do not edit here in isolation —
  any change is a brand-identity spec change.
-->

<img src="./site/public/mascot/companion-wave.svg" width="200"> <img src="./site/public/wordmark.svg" width="600">

[![Go Report Card](https://goreportcard.com/badge/github.com/nathanbrophy/glacier)](https://goreportcard.com/report/github.com/nathanbrophy/glacier) [![Static Badge](https://img.shields.io/badge/github-Read_The_Docs-teal?style=plastic&logo=github&logoColor=teal&labelColor=white)](https://nathanbrophy.github.io/glacier/) [![Static Badge](https://img.shields.io/badge/Go-Go_Docs-blue?style=plastic&logo=go)](https://pkg.go.dev/github.com/nathanbrophy/glacier)




Glacier is a Go framework that handles the plumbing so you can focus on what's yours. Like a glacier that shapes the landscape beneath the surface, Glacier is stable, deep, and predictable about the messy parts: argument parsing, configuration layering, lifecycle and signal handling, mock-driven testing, and HTTP transport faking. You write the logic. Glacier handles the rest.

## Status

Glacier is in early design. The repo currently holds the development lifecycle and the brand identity. Code lands as component specs are accepted.

- [`specs/`](specs/) — the source of truth. Every change is a spec first.
- [`specs/0000-spec-process.md`](specs/0000-spec-process.md) — how Glacier is built.
- [`specs/0001-brand-identity.md`](specs/0001-brand-identity.md) — what Glacier looks and feels like.
- [`CLAUDE.md`](CLAUDE.md) — the rules.

## Glacier SDK

The Glacier SDK is a CLI binary, `glacier`, that demonstrates the framework in production. Nine commands cover the developer day.

### Install

```sh
go install github.com/nathanbrophy/glacier/cmd/glacier@latest
```

### What ships

| Command | Description |
|---|---|
| `glacier init` | Scaffold a new Glacier project |
| `glacier new` | Add a package, command, or option to an existing project |
| `glacier generate` | Run code generators (cli, mock, httpmock) |
| `glacier lint` | gofmt, vet, staticcheck, and Glacier-specific lints |
| `glacier test` | Live status panel and summary |
| `glacier vibe` | Glacier vibes animation |
| `glacier version` | Print version; `--check` compares against latest |
| `glacier explain` | Explain a marker, exit code, or config key |
| `glacier completions` | Shell completions (bash, zsh, fish, pwsh) |

### Why dogfood

The SDK is the framework's longest-running integration test. Every Glacier package is exercised by at least one SDK command. A Lynx-owned coverage row fails CI if any package falls out of use. Bugs in the framework surface immediately, before they reach framework consumers.

[Read the SDK reference](https://nathanbrophy.github.io/glacier/sdk/)

## The Promise

When you use Glacier, you should be able to say each of these truthfully:

1. *"I'm only writing what's mine."*
2. *"I trust the defaults."*
3. *"The error tells me what to do next."*
4. *"Tests are easy because the framework helps."*

Every component spec is reviewed against these four statements. If a design doesn't deliver them, the design is wrong.

## License

License will be selected when the first code spec lands.
