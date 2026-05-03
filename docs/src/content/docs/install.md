---
title: Install
description: Three ways to get the lfg binary on your machine.
---

## One-liner (recommended)

```sh
curl -fsSL https://raw.githubusercontent.com/ptmaroct/lfg/main/install.sh | sh
```

Drops `lfg` in `/usr/local/bin` (or `~/.local/bin` when
`/usr/local/bin` isn't writable). After install, run `lfg` to launch
the TUI.

### Pin a version or change the prefix

```sh
LFG_VERSION=v0.3.0 LFG_PREFIX=$HOME/.local sh install.sh
```

| Var | Default | Notes |
|---|---|---|
| `LFG_VERSION` | latest GitHub release | Must match a published tag, e.g. `v0.3.0`. |
| `LFG_PREFIX` | `/usr/local` or `$HOME/.local` | Binary lands in `<prefix>/bin/lfg`. |

The script picks the right tarball for your `uname -s` / `uname -m`
combo (Darwin/Linux × amd64/arm64) from the goreleaser release.

## From source

Requires Go 1.25+:

```sh
git clone https://github.com/ptmaroct/lfg
cd lfg
go build -o lfg ./cmd/lfg
./lfg
```

`make build` is the same thing wrapped in a Makefile target, plus it
sets the version ldflags so `lfg version --verbose` returns the
correct git SHA.

## Self-update

Already have lfg? Skip the curl step:

```sh
lfg update
```

Pulls the latest release from GitHub via
[minio/selfupdate](https://github.com/minio/selfupdate). See the
[self-update guide](/lfg/guides/update/) for details.

## Verify

```sh
lfg version --verbose
```

Should print the version, git SHA, build date, Go version, and
target platform.

## Uninstall

`lfg` doesn't install a daemon, doesn't write to system directories,
and doesn't drop config until you launch it for the first time.
Uninstalling is two `rm`s:

```sh
rm $(command -v lfg)         # binary
rm -rf ~/.config/lfg         # state, logs, key, config
```

The `# lfg-managed PATH` block in your shell rc files can be deleted
by hand — search for `# lfg-managed PATH` in `~/.bashrc`, `~/.zshrc`,
and any fish config.
