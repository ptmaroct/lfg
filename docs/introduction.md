---
title: Introduction
description: What lfg is, what it isn't, and the design ideas that shaped it.
---

`lfg` is a single Go binary that launches an interactive TUI for
bootstrapping a fresh dev machine. It's open source, MIT-licensed,
and runs on macOS and Linux (Windows is on the long-term roadmap).

## What lfg does

- **Bootstraps developer tooling** — Homebrew, mise, node, bun, go,
  CLIs, AI coding harnesses, and any custom package you describe in a
  TOML preset.
- **Picks a curated subset** — bundles + a tree picker so you don't
  have to remember every tool by name.
- **Detects what's already there** — concurrent binary + version probe
  on launch; anything already on PATH lands in an *Already installed*
  group instead of being reinstalled.
- **Resolves live versions** — npm, brew, and Node.js registries hit
  asynchronously per tool with a 3 s timeout, so the picker shows you
  what's current without blocking the UI.
- **Augments your shell rc** — adds a fenced `# lfg-managed PATH` block
  to every shell rc it finds (`.bashrc`, `.zshrc`, fish), idempotent.
- **Backs up your dotfiles** — `lfg backup` produces a `tar.gz` or
  `tar.age` (encrypted) of your shell, editor, AI-tool, and SSH
  configs. Private keys stay on-device unless you opt in.
- **Self-updates** — `lfg update` pulls the latest release from
  GitHub.

## What lfg is not

- A package manager — it wraps brew/mise/npm rather than reinventing
  them.
- A dotfile manager — chezmoi is great at that; backup/restore in
  `lfg` is for the day-zero archive, not ongoing sync.
- A configuration management system — Nix, Ansible, and friends are
  the right answer for fleets. `lfg` is for one human, one laptop,
  Friday afternoon.

## Status

v0.3 is the latest stable: real installers, PostInstall hooks, async
live version resolver, `lfg backup` / `lfg doctor` / `lfg update`,
four themes. v0.4 ships the `terminal-essentials` bundle and renames
`dev-tools` to `ai-harnesses` — currently on the beta channel via
`brew install ptmaroct/tap/lfg-beta`. Snapshot tests still see
deterministic mock data so test runs never touch your system.

The roadmap (sync, SSH fleet management, macOS `defaults` wizard) is
on the [Roadmap page](/project/roadmap).
