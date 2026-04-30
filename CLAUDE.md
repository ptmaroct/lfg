# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`lfg` is an open-source TUI bootstrap CLI for new dev machines. **v0.1
ship is real**: installers (brew/apt/mise/npm/custom) execute live,
detect pass probes binaries, `lfg backup` produces tar/tar.age, `lfg
doctor` runs environment checks, `lfg update` self-updates from
GitHub releases. Snapshot tests (`make test`) and the `cmd/snap`
helper still see deterministic mock data — never the host system.

Cobra-based subcommand surface: `lfg`, `lfg apply`, `lfg backup`,
`lfg doctor`, `lfg version`, `lfg update`. Default `lfg` (no args)
launches the TUI. Theme persists in `~/.config/lfg/state.json`.

## UI rule — always use Charm components

This project lives in the Charm ecosystem and ships `bubbletea`, `huh`,
`bubbles`, `lipgloss` already. **Use them. Do not hand-roll widgets when
a Charm component exists.**

| Need | Use | Don't write |
|------|-----|-------------|
| yes/no decision | `huh.NewConfirm` | custom Y/N button row |
| pick one of N | `huh.NewSelect` | custom cursor list |
| pick many | `huh.NewMultiSelect` | custom checkbox list |
| text input | `huh.NewInput` / `bubbles/textinput` | custom prompt |
| password / secret | `huh.NewInput().EchoMode(huh.EchoPassword)` | masked rolled-by-hand |
| filterable list | `bubbles/list` | custom delegate from scratch |
| spinner | `bubbles/spinner` | dot animation by hand |
| progress bar | `bubbles/progress` | block-string interpolator |
| stopwatch / timer | `bubbles/stopwatch` / `bubbles/timer` | time.Since loop |
| scrolling log | `bubbles/viewport` | manual line-tail slice |
| key help footer | `bubbles/help` (or our `KeyHint`/`HintLine`) | from scratch |
| paginator | `bubbles/paginator` | counter widget |

Style/theming goes through `lipgloss` + `HuhTheme(palette)` — no raw
ANSI escapes, no inline ad-hoc color constants.

When the design needs something Charm doesn't ship (e.g. the welcome
hero, the tactical stats row, the tree picker rows), build it on top of
`lipgloss` styles — but the *interactive* primitives must be Charm.
This keeps keybindings, theming, focus management, and a11y consistent
across the app.

If you find yourself reinventing buttons, checkboxes, selects, or text
inputs: stop and use `huh`.

## Common commands

```sh
make build           # ./lfg binary
make run             # build + run with default theme
make test            # snapshot tests across 6 widths × 3 themes (~0.6s)
make snap-update     # regenerate testdata/*.golden after intentional UI change
make preview                       # page through every golden in less -R
make preview ARGS=welcome          # filter goldens by substring
make preview ARGS="bundles xl"     # multi-arg = AND filter
make widths          # static PNGs per width (needs `brew install charmbracelet/tap/freeze`)
make demo            # record full flow as gif (needs `brew install vhs`)
make docker          # build container image (Ubuntu + linuxbrew)
make docker-run      # launch TUI in container
```

Single test:
```sh
go test ./internal/tui -run TestSnapshot_Welcome/welcome_lfg_md_100x30
```

The repo's Go module path is `github.com/anuj/lfg` but the GitHub home is
`ptmaroct/lfg` — rename the module before public OSS launch.

## Architecture

### Screen state machine (`internal/tui/app.go`)

`Model` holds a `screen` enum + a child model per screen. The root
`Update` dispatches messages to the active child; child models request
transitions by returning a `transitionMsg` cmd. The `transition` method
re-instantiates the destination child model with current selection state
so each screen always sees fresh, correct data.

**Selection state** flows screen → screen via two maps on the root model:
- `selectedBundleIDs map[string]bool` (set in bundle picker)
- `selectedTools map[string]bool` keyed `<bundleID>/<toolName>`

When adding a screen: add to the `screen` enum, add a child field on
`Model`, register dispatch in `Update`/`forwardSize`/`View`/`transition`,
and emit `goTo(screenX)` from wherever you trigger the transition.

### Theming (`internal/tui/theme.go`)

`Palette` is the abstract palette every screen uses (Primary, Accent,
Success, Muted, Gradient, etc.). Three palettes ship: `lfg`, `dracula`,
`catppuccin`. `HuhTheme(p)` derives a `*huh.Theme` from a palette by
cloning `huh.ThemeCharm()` and overriding select fields.

**The huh `Group.Base` border is intentionally hidden** — the visible
"vertical bars on the left" bug came from huh's group border showing
inside our outer Frame. Don't re-enable that border without removing
the outer Frame card too.

Adding a theme: extend `PaletteFor` with a new case + add the name to
the `flag` validation in `cmd/lfg/main.go` + add to `screens_test.go`'s
welcome theme loop so it gets snapshot coverage.

### Layout (`internal/tui/layout.go`)

`Frame(palette, w, h, subtitle, inner, footer, compactTitle)` is the
single source of chrome. Every screen's `View(w, h)` ends in a `Frame`
call. Inner widgets get wrapped in a card (rounded border, fixed
`canvasW-12` width) before being placed via `lipgloss.PlaceHorizontal`.

**Why the card wrapper:** without it, huh forms render flush-left
within their centered bounding box (long lines stretch the bbox, short
lines stay glued to the left). The card gives every screen a uniform
rectangle to center.

`compactTitle` is `height < 22` — falls back from figlet banner to a
single-line gradient `lfg` string. Don't break this contract; xs/sm
test widths depend on it.

### Title (`internal/tui/title.go`)

Renders an ASCII figlet via `go-figure` then per-char gradient via the
hand-rolled `blend1D`. `lipgloss` v1.1.0 doesn't have built-in 1D color
blending; v2 does. If we upgrade lipgloss to v2, replace `blend1D` with
the upstream call.

**Font is `standard`** — switching fonts is a one-line change in
`title.go:25`. Bundled options: `standard`, `slant`, `big`, `block`,
`doom`, `larry3d`, `roman`, `script`, `shadow`, `term`, etc. `slant`
made `lfg` look slashy; `big` had `g` descender bleed; `standard` is
the cleanest balance for short brand names.

Every line is right-padded to `maxW` so downstream `Align(Center)` /
`PlaceHorizontal` treats the figlet block as one rectangle. Removing
that padding will break centering on the next render pass.

### Snapshot testing (`internal/tui/screens_test.go`)

**Don't use teatest's byte-stream capture for snapshots** — it returns
the full ANSI escape stream including alt-screen sequences, useless
for diffing. Instead, the test bypasses `tea.Program` entirely:

1. `New(theme)` → root Model
2. Drive via direct `m.Update(...)` calls with synthetic key messages
3. Call `m.View()` to capture the rendered string
4. Diff against `testdata/<name>.golden`

The `cmd` returned by `Update` is resolved exactly once (via `cmd()`)
to surface synchronous transitions. Async cmds (timers, ticks) don't
fire — fine for our case because every screen renders meaningfully on
first paint.

Update flag: prefer `-update` (the `flag.Lookup("update")` lookup
catches teatest's reserved flag). `-update-snap` is an alias.

When you change UI: `make test` shows diff. If intentional,
`make snap-update`, then commit `internal/tui/testdata/*.golden` so
reviewers can eyeball the diff in the PR.

### Preset data (`internal/preset/preset.go`)

Bundles + tools are **hardcoded for the prototype**. v0.1 will fetch
TOML from `raw.githubusercontent.com/<org>/presets/main/default.toml`.
The `Tool.Installed` + `Tool.Version` fields are also hardcoded fake
data so the picker has something to show — these will come from a real
detect pass (`internal/detect/`) once installers exist.

## Docker

`Dockerfile` is multi-stage: Go build → Ubuntu 24.04 + Homebrew. The
brew install runs as non-root user `dev` (brew refuses root on Linux).
Image is ~1.05GB because of the brew toolchain.

When iterating UI in container: prefer Mode 2 from README (live source
mount via `golang:1.26-bookworm`) — no rebuild needed per change.

## Plan + roadmap

The agent plan that drove the design lives outside the repo at
`/Users/anuj/.claude/plans/what-are-some-of-eventual-coral.md` (also
copied to `plan.md` which is gitignored). README has the v0.1 → v1
roadmap. v0.2 = GitHub auth + remote sync. v0.3 = SSH + macOS defaults.
Anything not in v0.1 should not land without revisiting that plan.

## Things to avoid

- Re-enabling huh `Group.Base` border without removing the outer Frame card.
- Per-line `Align(Center)` on multi-line content (figlet, log tail) — use
  `PlaceHorizontal` so blocks center as rectangles.
- Touching `internal/tui/testdata/` by hand — always regenerate via
  `make snap-update` so all themes/widths stay in sync.
- Wiring real subprocesses into snapshot tests — `tui.New(theme)` uses
  `mockProgressRunner`; CLI startup passes `installer.Run`. Don't
  conflate the two paths.
