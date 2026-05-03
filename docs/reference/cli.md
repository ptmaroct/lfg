---
title: CLI commands
description: Every subcommand and persistent flag.
---

```text
lfg [command] [flags]
```

Default (no subcommand) launches the TUI.

## Commands

| Command | Description |
|---|---|
| `lfg` | Launch the interactive TUI. |
| `lfg apply <bundle...>` | Headless install of one or more bundles. |
| `lfg backup` | Snapshot dotfiles + configs into `tar` or `tar.age`. |
| `lfg backup --restore <path>` | Extract a previous backup. |
| `lfg export` | Save the current preset to `~/lfg-preset-<date>.toml`. |
| `lfg doctor` | Diagnose environment readiness. |
| `lfg version` | Print the version. |
| `lfg version --verbose` | Print version + git SHA + build date + Go version + platform. |
| `lfg update` | Self-update from GitHub releases. |

## Persistent flags

These work on every command:

| Flag | Default | Description |
|---|---|---|
| `--theme <name>` | `lfg` | Palette: `lfg`, `dracula`, `catppuccin`, `colorblind`. |
| `--config <path>`, `-c` | built-in | Custom preset TOML (file path or `http(s)://` URL). |
| `--bg <mode>` | `auto` | Force terminal background: `auto`, `dark`, `light`. Also reads `LFG_BG`. |
| `--debug` | off | Verbose log → `~/.config/lfg/logs/debug-<ts>.log`. |
| `--dry-run`, `-n` | off | Walk the flow without exec'ing anything. |
| `--help`, `-h` | — | Show help for the command. |

## Per-command flags

### `lfg apply`

| Flag | Description |
|---|---|
| `<bundle...>` | One or more bundle IDs (positional). |
| `-n` | Print planned commands, exec nothing. |

### `lfg backup`

| Flag | Default | Description |
|---|---|---|
| `--encrypt` | `true` | Produce `.tar.age` (encrypted). `--encrypt=false` → `.tar.gz`. |
| `--include-ssh-keys` | off | Include private keys (only honoured when `--encrypt`). |
| `--restore <path>` | — | Extract a previous backup. |
| `--key <path>` | `~/.config/lfg/key.txt` | Override key location for `.tar.age`. |
| `--force` | off | Overwrite existing files on restore. |
| `-n` | off | Print source counts and would-be filename, write nothing. |

### `lfg export`

| Flag | Default | Description |
|---|---|---|
| `-o <path>` | `~/lfg-preset-YYYY-MM-DD.toml` | Output path. |
| `-n` | off | Print path it would write, no file created. |

### `lfg update`

| Flag | Description |
|---|---|
| `--check` | Show available version, don't apply. |
| `--version <tag>` | Pin to a specific release tag. |

### `lfg doctor`

| Flag | Description |
|---|---|
| `--verbose` | Add resolved binary path, version, and timing per check. |

## Exit codes

- `0` — success
- non-zero — first failure surfaces; check stderr for details

For installer failures during `lfg apply`, the exit code is the
shell exit code of the failing step.
