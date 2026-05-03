---
title: Bundles & tools
description: How lfg organises packages and how the picker behaves.
---

A **bundle** is a named group of related tools. A **tool** has a name,
a homepage, an install command (per OS), and a backend that tells lfg
how to run it (`brew`, `apt`, `mise`, `npm`, `curl`, `skills`, or
`custom`).

## Built-in bundles

The default preset ships with bundles for a typical dev box:

- **barebones** — Homebrew, mise, node-lts, pnpm, bun, yarn, python, uv.
- **terminal-essentials** — interactive shell quality-of-life: zoxide,
  ripgrep, lazygit, tailscale, tmux, plus `ghostty` (macOS only, brew
  cask) and `zsh` (Linux only, since recent macOS already ships a
  current zsh).
- **ai-harnesses** — AI coding CLIs (Claude Code, Codex, OpenCode, Droid).
- **skills** — cross-harness AI skills installed via `npx skills add`.
  Gated until `node-lts` is selected.
- **mcp** — official Model Context Protocol servers for any MCP-aware
  harness.

You can replace this entirely with a custom preset — see
[Custom presets](/guides/presets).

![tree](/screens/tree.png)

## Tree picker behaviour

| Marker | Meaning |
|---|---|
| `[●]` | mandatory — can't be unchecked |
| `[✓]` | selected |
| `[ ]` | not selected |
| `→ v1.2.3` | live latest version (looked up from upstream registry) |
| `ALREADY INSTALLED` | bucket for tools detected on `PATH` already |

Live versions resolve asynchronously per tool with a 3 s timeout —
the picker renders immediately with the parsed `@version` from the
install command and updates in place once the registry responds.

## Tool info card

![info](/screens/info.png)


Press `i` on any tool to open a card showing:

- description
- homepage / repo
- the exact install command for your OS
- any `post_install` steps that will run after the main install

Useful for sanity-checking what a bundle is about to do before you
hit confirm.

## Source backends

| Source | What it does |
|---|---|
| `brew` | `brew install <name>` (macOS + linuxbrew on Linux) |
| `apt` | `sudo apt-get install -y <name>` (Linux) |
| `mise` | `mise use -g <plugin>@<ver>` for language runtimes |
| `npm` | `npm i -g <pkg>` |
| `curl` | arbitrary `curl … \| sh` install scripts |
| `skills` | `npx skills add <url>` (requires node-lts) |
| `custom` | the `install_mac` / `install_linux` strings as-is |

Bootstrap-able backends (brew, mise, node) are auto-installed at the
start of a run when at least one selected tool needs them.
