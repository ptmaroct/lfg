---
title: Roadmap
description: What's shipped, what's next, what's deferred.
---

## v0.1 — MVP (shipped)

- TUI skeleton with all screens
- Animated gradient logo, tactical bulletin chrome
- Themes, swappable via `--theme` or `Ctrl+T`
- Tree picker with bundle/tool selection
- Quit-confirm dialog
- Real package installers (brew, apt, mise, npm, custom)
- Binary-present + version detection
- Real `lfg backup` → tar + optional age encrypt
- `lfg doctor` env checks
- `lfg update` self-update
- goreleaser + `install.sh`

## v0.3 — installers + live data (shipped)

- PATH augmentation between steps
- PostInstall hook chain per tool
- Async live version resolver (npm / brew / nodejs registries)
- `cmd/snap` helper for deterministic screenshot generation
- Cleaned-up Docker images (with + without prebaked brew)

## v0.2 — sync (planned)

- GitHub device-flow auth
- Dedicated `lfg-config` repo per user
- `lfg sync` / `lfg apply` / `lfg diff`
- Dotfile restore (not just backup)

## v0.3 — SSH + macOS (planned)

- `lfg ssh list` (wishlist-powered)
- `lfg ssh add-device` (fleet pubkey push)
- macOS `defaults` + `systemsetup` wizard

## Deferred / v1+

- Hosted paid sync tier
- Plugin system for user-authored install recipes
- Teams / shared configs
- Windows support

If something here matters to you, +1 the corresponding GitHub issue
or open one. Roadmap is shaped by what people are actually trying
to use.
