---
title: Themes
description: Four built-in palettes and how to switch between them.
---

`lfg` ships four palettes. Pick one with `--theme=<name>` at launch,
or cycle live with `Ctrl+T`. The active theme is shown top-right of
every screen, and your choice persists in `~/.config/lfg/state.json`.

| LFG (default) | Dracula | Catppuccin |
|---|---|---|
| ![lfg](/lfg/screens/welcome.png) | ![dracula](/lfg/screens/welcome-dracula.png) | ![catppuccin](/lfg/screens/welcome-catppuccin.png) |

## Built-in palettes

| Name | Vibe | Accent stops |
|---|---|---|
| `lfg` (default) | brand — pink → violet → emerald gradient | Tailwind `-400` for AA contrast on near-black terminals |
| `dracula` | the classic dark | matches [draculatheme.com](https://draculatheme.com) |
| `catppuccin` | warm pastels | Mocha variant |
| `colorblind` | IBM blue / orange / magenta | designed for deuteranopia / protanopia |

```sh
lfg --theme=dracula
lfg --theme=catppuccin
lfg --theme=colorblind
```

## Live cycle

Press `Ctrl+T` from any screen to rotate through the palettes. The
TUI re-renders in place; nothing else changes. The new selection is
written to state on the next idle tick, so your next launch picks up
where you left off.

## Why `-400` accents

The default `lfg` palette intentionally targets the Tailwind `-400`
stops (pink-400 / violet-400 / emerald-400). The earlier `-500` /
`-600` versions only scored ~3:1 contrast against near-black
terminals — small `01` digits and `▸` cursor glyphs at body weight
came out near-invisible. `-400` lands at ~5:1 against black while
still readable on white.

## Adding a theme

If you want a new palette in the binary itself:

1. Add a case to `PaletteFor` in `internal/tui/theme.go`.
2. Add the name to the `validateTheme` switch in
   `internal/cli/root.go`.
3. Add it to the welcome theme loop in
   `internal/tui/screens_test.go` so it gets snapshot coverage.
4. Run `make snap-update` to generate goldens.

PR welcome — accessibility palettes especially.
