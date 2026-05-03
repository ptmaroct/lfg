---
title: State files
description: What lfg writes under ~/.config/lfg/ and what each file is for.
---

`lfg` keeps everything under `$XDG_CONFIG_HOME/lfg/` (defaults to
`~/.config/lfg/`).

```
~/.config/lfg/
├── state.json              persistent UI state (theme, last-used bundles)
├── key.txt                 age key for backup encryption (chmod 600)
└── logs/
    ├── install-<ts>.log    full transcript of the most recent install run
    └── debug-<ts>.log      verbose log when --debug is passed
```

## state.json

```json
{
  "theme": "dracula",
  "version": "v0.4.0",
  "lastBundles": ["barebones", "terminal-essentials", "ai-harnesses"]
}
```

Written on theme cycle (`Ctrl+T`) and on install completion. Safe to
delete — `lfg` regenerates with defaults on next launch.

## key.txt

[age](https://github.com/FiloSottile/age) X25519 keypair. Created on
first `lfg backup` if missing. **Back this file up separately** —
without it, encrypted backups are unrecoverable.

```
# created by lfg on 2025-01-15
# public key: age1...
AGE-SECRET-KEY-1...
```

Permissions are forced to `600` on creation. If you symlink to a
backup location (e.g. a 1Password ssh agent) lfg honours the link.

## logs/

Rotated by run, never auto-deleted — `lfg` won't delete logs you
might want for debugging. Clean up by hand:

```sh
find ~/.config/lfg/logs -mtime +30 -delete
```

`install-*.log` is always written when an install runs.
`debug-*.log` is only written when `--debug` is passed; it includes
process spawn/exit diagnostics that aren't in the user-facing log.

## Migration

There is no migration system yet — schema changes between minor
versions are handled by overwriting `state.json` with sensible
defaults. The on-disk format is intentionally minimal so this stays
cheap.
