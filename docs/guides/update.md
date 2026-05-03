---
title: Self-update
description: Pull the latest release without re-running curl | sh.
---

```sh
lfg update           # update to latest release
lfg update --check   # show available version, don't apply
```

Pulls the matching tarball for your OS/arch from the GitHub releases
page, verifies the SHA-256, and atomically replaces the running
binary in place via [minio/selfupdate](https://github.com/minio/selfupdate).

## Pin a version

```sh
lfg update --version=v0.3.0
```

Useful for downgrading after a regression, or pinning a fleet to a
known-good release.

## Beta channel

Beta releases ship from `develop` ahead of stable. Two ways to track
them:

```sh
brew install ptmaroct/tap/lfg-beta     # parallel install, separate binary
lfg update --version=v0.4.0-beta.1     # one-off pin, replaces running binary
```

`lfg update` (no flag) only follows the **latest stable** release —
beta tags never get auto-pulled. Use the tap formula if you want
betas to update transparently with `brew upgrade`.

## How it works

1. Hits `api.github.com/repos/ptmaroct/lfg/releases/latest` (or the
   tag you passed) and reads the asset list.
2. Picks the right asset for your `runtime.GOOS` / `runtime.GOARCH`.
3. Downloads the tarball + checksums file, verifies the SHA.
4. Atomically swaps the new binary into place; on failure, restores
   the old one.

No daemon, no auto-update at launch — `lfg update` only runs when
you ask it to.

## CI / unattended

`lfg update` exits non-zero on any failure (network, checksum
mismatch, atomic-rename failure). Safe in cron:

```sh
# crontab entry — weekly Mon 04:00
0 4 * * 1 /usr/local/bin/lfg update >/var/log/lfg-update.log 2>&1
```

`--check` is a dry-run that prints the available version to stdout
without applying — wire it into a shell alias if you'd rather be
notified than auto-updated.
