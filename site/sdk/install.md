---
title: Install the Glacier SDK
---

# Install the Glacier SDK

Requires Go 1.25 or later (toolchain go1.26.0 recommended).

## Option 1: go install (recommended)

```sh
go install github.com/nathanbrophy/glacier/cmd/glacier@latest
```

`go install` places the binary in `$GOPATH/bin` (usually `~/go/bin` on Linux/macOS or `%USERPROFILE%\go\bin` on Windows). Make sure that directory is on your `PATH`.

## Option 2: build from source

```sh
git clone https://github.com/nathanbrophy/glacier.git
cd glacier
go build -o glacier ./cmd/glacier
```

Move the resulting binary somewhere on your `PATH`.

## Option 3: build a specific tagged release

```sh
git clone --branch v0.1.0 --depth 1 https://github.com/nathanbrophy/glacier.git
cd glacier
go build -o glacier ./cmd/glacier
```

Replace `v0.1.0` with the tag you want.

## Adding glacier to PATH

### macOS and Linux

```sh
export PATH="$PATH:$(go env GOPATH)/bin"
```

Add that line to `~/.bashrc`, `~/.zshrc`, or your shell's profile to make it permanent.

### Windows (PowerShell)

```powershell
$env:PATH += ";$(go env GOPATH)\bin"
```

To make it permanent, add the path through System Properties > Environment Variables, or add the line to your PowerShell profile (`$PROFILE`).

### Windows (Git Bash)

```sh
export PATH="$PATH:$(go env GOPATH)/bin"
```

Add to `~/.bashrc` or `~/.bash_profile`.

## Verifying the install

```sh
glacier version
```

Expected output (version number varies):

```
ʕ•ᴥ•ʔ glacier version
ʕ⌐■-■ʔ glacier v0.1.0
  go:    go1.26.0
  os:    linux/amd64
```

## Shell completions

Run once to enable tab completion.

**bash**

```sh
glacier completions bash > ~/.local/share/bash-completion/completions/glacier
# or load for the current session only:
source <(glacier completions bash)
```

**zsh**

```sh
glacier completions zsh > "${fpath[1]}/_glacier"
# then restart your shell or run:
compinit
```

**fish**

```sh
glacier completions fish > ~/.config/fish/completions/glacier.fish
```

**PowerShell**

```powershell
glacier completions pwsh >> $PROFILE
```

See [`glacier completions`](./commands/completions.md) for details on each install location.

## Upgrading

The SDK never auto-updates. When `glacier version --check` reports a newer release, upgrade with:

```sh
go install github.com/nathanbrophy/glacier/cmd/glacier@latest
```

## Configuration (optional)

The config file lives at `<UserConfigDir>/glacier/config.json`. The SDK runs fine with no config file. See [Configuration](./configuration.md) for all keys.
