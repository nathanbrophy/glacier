---
title: Install the Glacier SDK
---

# Install the Glacier SDK

`ʕ•ᴥ•ʔ` One command. Any platform.

```sh
go install github.com/nathanbrophy/glacier/cmd/glacier@latest
```

Requires Go 1.22 or later. The install command places the `glacier` binary in your `$GOPATH/bin`. Confirm it is on your `PATH`:

```sh
glacier version
```

## Supported platforms

| OS | Architecture |
|---|---|
| Linux | amd64, arm64 |
| macOS | amd64, arm64 |
| Windows | amd64 |

Binaries are cross-compiled in CI for every platform. All platforms are first-class.

## Shell completions

Pick your shell and run once. The completion script is generated from the live command tree, so it stays current with your installed version.

**bash**

```sh
glacier completions bash > ~/.local/share/bash-completion/completions/glacier
# or: source <(glacier completions bash)
```

**zsh**

```sh
glacier completions zsh > "${fpath[1]}/_glacier"
# Then restart your shell or run: compinit
```

**fish**

```sh
glacier completions fish > ~/.config/fish/completions/glacier.fish
```

**PowerShell**

```powershell
glacier completions pwsh >> $PROFILE
```

See [`glacier completions`](./commands/completions.md) for details on each shell's install location.

## Upgrading

The SDK never auto-updates. When `glacier version --check` reports a newer release, upgrade with the same install command:

```sh
go install github.com/nathanbrophy/glacier/cmd/glacier@latest
```

## Configuration

Optional. The config file lives at `<UserConfigDir>/glacier/config.json`. See [Configuration](./configuration.md) for the full key reference.
