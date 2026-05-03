---
title: Environment variables
description: Vars lfg reads at runtime, plus the install-time vars from install.sh.
---

## Runtime

| Var | Default | Effect |
|---|---|---|
| `LFG_CONFIG` | unset | Same as `--config` — path or `http(s)://` URL to a TOML preset. |
| `LFG_THEME` | `lfg` | Same as `--theme`. |
| `LFG_NO_COLOR` | unset | Disables ANSI styling. Honoured for accessibility / piping output. |
| `NO_COLOR` | unset | Standard `NO_COLOR` is honoured (same effect as `LFG_NO_COLOR`). |
| `LFG_DEBUG` | unset | Same as `--debug`. |
| `XDG_CONFIG_HOME` | `~/.config` | Where state lives — `$XDG_CONFIG_HOME/lfg/`. |

## Install-time (install.sh)

| Var | Default | Effect |
|---|---|---|
| `LFG_VERSION` | `latest` | Pin the install to a specific release tag (`v0.3.0`). |
| `LFG_PREFIX` | `/usr/local`, fallback `$HOME/.local` | Binary lands in `<prefix>/bin/lfg`. |

## Inherited from the system

`lfg` shells out to brew, mise, npm, etc. — those each respect their
own env (`HOMEBREW_NO_ANALYTICS`, `MISE_QUIET`, `NPM_CONFIG_REGISTRY`,
…). `lfg` doesn't override or wrap them; whatever you've set will
apply.

## TUI-specific

| Var | Effect |
|---|---|
| `TERM` | Used for terminal capability detection. The fancy gradient and figlet logo require a 256-colour terminal. |
| `COLORTERM` | If `truecolor` or `24bit`, the gradient blends across the full RGB range. |

If `TERM` reports a dumb terminal, `lfg` falls back to plain ASCII
without colour.
