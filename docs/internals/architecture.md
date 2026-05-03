---
title: Architecture
description: How the binary is structured and what each package owns.
---

| Layer | Choice |
|---|---|
| Language | Go 1.25+ |
| TUI framework | [bubbletea](https://github.com/charmbracelet/bubbletea) |
| Components | [bubbles](https://github.com/charmbracelet/bubbles) — spinner, progress, stopwatch |
| Forms | [huh](https://github.com/charmbracelet/huh) — select, multi-select, confirm |
| Styles | [lipgloss](https://github.com/charmbracelet/lipgloss) |
| Config format | TOML |
| CLI lib | [cobra](https://github.com/spf13/cobra) |
| Encryption | [age](https://github.com/FiloSottile/age) — backup payloads |
| Self-update | [minio/selfupdate](https://github.com/minio/selfupdate) |
| Packaging | [goreleaser](https://goreleaser.com) → tarballs + `install.sh` |

## Repo layout

```
lfg/
├── cmd/
│   ├── lfg/main.go         thin entry — calls cli.Execute()
│   └── snap/main.go        screen → ANSI text helper for screenshots
├── internal/
│   ├── cli/                cobra subcommands (root, apply, backup, doctor, ...)
│   ├── preset/             bundle + tool data (built-in default + parser)
│   ├── detect/             binary + version probe (concurrent, timeout-bounded)
│   ├── installer/          brew/apt/mise/npm/custom + streaming exec
│   ├── backup/             tar + age snapshot pipeline
│   ├── doctor/             environment readiness checks
│   ├── state/              ~/.config/lfg/state.json
│   ├── version/            ldflags-injected build metadata
│   └── tui/
│       ├── app.go          screen state machine + global hotkeys
│       ├── theme.go        4 palettes + huh theme builder
│       ├── layout.go       Frame(): outer chrome, centering
│       ├── title.go        figlet logo + gradient sweep
│       ├── welcome.go      animated hero + numbered actions
│       ├── tree_picker.go  collapsible bundle/tool tree
│       ├── confirm.go      stats row + huh.Confirm
│       ├── progress.go     channel-driven log tail wired to installer
│       ├── done.go         next-steps card
│       ├── backup.go       huh.Confirm (encrypt) → spinner → result
│       └── quit_confirm.go huh.Confirm dialog (`q` from anywhere)
├── .goreleaser.yaml
├── .github/workflows/
└── install.sh
```

## Screen state machine

`internal/tui/app.go::Model` holds a `screen` enum + a child model per
screen. The root `Update` dispatches messages to the active child;
child models request transitions by returning a `transitionMsg` cmd.
The `transition` method re-instantiates the destination child model
with current selection state so each screen always sees fresh, correct
data.

Selection state flows screen → screen via two maps on the root model:

- `selectedBundleIDs map[string]bool` (set in bundle picker)
- `selectedTools map[string]bool` keyed `<bundleID>/<toolName>`

Adding a screen: extend the `screen` enum, add a child field on
`Model`, register dispatch in `Update`/`forwardSize`/`View`/`transition`,
and emit `goTo(screenX)` from wherever you trigger the transition.

## Layout & chrome

`Frame(palette, w, h, subtitle, inner, footer, compactTitle)` is the
single source of chrome. Every screen ends in a `Frame` call. Frame
builds a fully closed bulletin box: top corners + side rules + bottom
corners.

`CanvasW(width)` is the single source of truth for inner canvas
width — every screen calls it instead of recomputing the formula
inline. Bounds: 56 min, 100 max.

`compactTitle` is `height < 22` — falls back from figlet banner to a
single-line gradient `lfg` string. xs/sm test widths depend on this
contract.
