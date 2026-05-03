---
title: Quick start
description: Launch the TUI, install a bundle, take a backup — under five minutes.
---

```sh
lfg                # launch TUI
lfg --theme=dracula
lfg apply barebones                # headless install of one bundle
lfg backup                         # snapshot dotfiles + configs
lfg --config ./my-preset.toml      # load custom bundles
```

`--dry-run` (or `-n`) walks every flow without exec'ing anything. Theme
persists in `~/.config/lfg/state.json`.

![welcome](/screens/welcome.png)

## The TUI flow

1. **Welcome** — pick install or backup. Logo gradient sweeps live.
2. **Tree picker** — expand a bundle (`→`), toggle individual tools
   (`space`), or toggle a whole bundle. `a` toggles everything; `i`
   opens a tool info card.
3. **Confirm** — summary of total to install, already present, source
   breakdown (brew · mise · npm · curl · skills).
4. **Progress** — gradient bar, spinner, per-task log tail, elapsed
   timer.
5. **Done** — reload-shell command + next-steps card.

## Backup flow

Welcome → Backup → encrypt y/n → spinner → `.tar` (or `.tar.age`)
result with key-backup reminder. Press `q` for quit-confirm,
`Ctrl+T` to cycle theme.

## Headless usage

If you already know which bundles you want, skip the TUI:

```sh
lfg apply barebones dev-tools     # multiple bundles, no prompts
lfg apply -n barebones            # print what it would run
```

`lfg apply` returns non-zero on installer failure, suitable for CI or
provisioning scripts.
