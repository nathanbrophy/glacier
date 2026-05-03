---
title: Glacier SDK
---

# Glacier SDK

<div class="glacier-sprite-accent">
  <MascotSprite state="wave" :size="96" />
</div>

ʕ•ᴥ•ʔ One command to install. Nine to cover the developer day.

The Glacier SDK is a CLI binary, `glacier`, built on every Glacier framework package. It scaffolds new projects, runs code generators, lints, tests, and explains the framework to you as you go. It is also the framework's longest-running integration test: every package Glacier ships is exercised by at least one SDK command, and a CI row fails if any package falls out of use.

## What ships

### CREATE

| Command | Description |
|---|---|
| [`glacier init`](./commands/init.md) | Scaffold a new Glacier project with signals, banner, version, completions, and OTEL wired from day one |
| [`glacier new`](./commands/new.md) | Add a package, command, or functional-option constructor to an existing project |

### DEVELOP

| Command | Description |
|---|---|
| [`glacier generate`](./commands/generate.md) | Run all three Glacier code generators (cli, mock, httpmock) concurrently |
| [`glacier lint`](./commands/lint.md) | gofmt, go vet, staticcheck, and six Glacier-specific lints with an optional auto-fix pass |
| [`glacier test`](./commands/test.md) | Wrap `go test -json` with a live status panel, a color-coded summary, and benchmark regression gating |

### INSPECT

| Command | Description |
|---|---|
| [`glacier version`](./commands/version.md) | Print version, Go toolchain, and OS/arch; `--check` fetches the latest release from GitHub |
| [`glacier explain`](./commands/explain.md) | Print a boxed explanation for any marker, exit code, or config key |

### UTILITY

| Command | Description |
|---|---|
| [`glacier vibe`](./commands/vibe.md) | Animated polar-bear banner with rotating tips; static fallback on non-TTY |
| [`glacier completions`](./commands/completions.md) | Print a shell-completion script for bash, zsh, fish, or PowerShell |

## Quickstart

Install the binary:

```sh
go install github.com/nathanbrophy/glacier/cmd/glacier@latest
```

Scaffold a project:

```sh
glacier init my-app --yes
cd my-app
```

Run the generators and tests to confirm the scaffold is wired:

```sh
glacier generate
glacier test
```

That is a working Glacier app. The `main.go` is six lines; all wiring lives in the generated `zz_generated_cli.go`.

## Where to go next

- [Install](./install.md) - per-platform install instructions, PATH setup, and shell completions
- [Commands](./commands/index.md) - all nine commands with flag tables and examples
- [Configuration](./configuration.md) - config file keys, environment variables, exit codes
