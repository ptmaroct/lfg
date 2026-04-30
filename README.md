# lfg

> Make a new dev machine feel like home in minutes, not hours.

`lfg` is an opinionated, open-source TUI that bootstraps a fresh Mac or
Linux dev box — installs the tools you pick, restores your dotfiles, and
(eventually) distributes your SSH identity to the servers you already
trust. One `curl | sh` away on day one; one command to backup what you
have on day N.

**Status:** early UX prototype. No installers wired yet — every screen
runs, every keybind works, but the "install" step is mocked so the UX
can be tuned before touching your system.

![welcome](assets/welcome.png)

## Why

Setting up a fresh machine eats a full evening:

1. Install Homebrew, mise, node, bun, go, …
2. Pull your dotfiles from somewhere.
3. Copy your SSH config, regenerate keys, add them to every server.
4. Tune a dozen macOS preferences.
5. Miss two of them, notice next Tuesday.

Existing tools solve parts of this (chezmoi, Omakub, dotbot, Nix). None
do the whole thing with a friendly interactive UI. `lfg` wraps the
best-in-class primitives — Homebrew, mise, chezmoi, age — in a
[Charm](https://charm.sh) TUI so the first five minutes on a new box
feel great.

## Screens

**Welcome** — animated gradient `LFG` logo, numbered actions.

![welcome](assets/welcome.png)

**Tree picker** — bundles + tools as a collapsible tree. Status dots
(● all · ◐ partial · ○ none), tabular ID / NAME / STATE / COUNT layout.

![tree](assets/tree.png)

**Themes** — three built-in palettes, cycle live with `Ctrl+T`.

| LFG (default) | Dracula | Catppuccin |
|---|---|---|
| ![lfg](assets/welcome.png) | ![dracula](assets/welcome-dracula.png) | ![catppuccin](assets/welcome-catppuccin.png) |

## Flow

1. **Welcome** — pick install or backup. Logo gradient sweeps live.
2. **Tree picker** — expand a bundle (→), toggle individual tools
   (space), or toggle a whole bundle. `a` toggles everything.
3. **Confirm** — telemetry-style summary: total to install, already
   present, source breakdown (brew · mise · custom).
4. **Progress** — gradient bar, spinner, log tail, elapsed timer.
5. **Done** — next-steps card.

Backup flow: `huh.Confirm` for encrypt y/n → spinner → `.tar[.age]`
result with key-backup reminder.

Press `q` from any screen → quit-confirm dialog.
Press `Ctrl+T` from any screen → cycle theme.

## Install (prototype)

Requires Go 1.22+.

```sh
git clone https://github.com/ptmaroct/lfg
cd lfg
go build -o lfg ./cmd/lfg
./lfg
```

Once the MVP lands, this becomes:

```sh
curl -fsSL https://raw.githubusercontent.com/lfg-cli/lfg/main/install.sh | sh
```

## Usage

```sh
./lfg                      # launch TUI (default theme)
./lfg --theme=lfg          # pink → purple → mint
./lfg --theme=dracula      # classic Dracula
./lfg --theme=catppuccin   # Catppuccin Mocha
```

### Key bindings

| Key | Action |
|---|---|
| `↑ ↓ j k` | move |
| `→ ←` | expand / collapse (tree) |
| `space` `x` | toggle option |
| `a` | toggle all (tree) |
| `enter` | confirm / continue |
| `1`–`3` | jump to action (welcome) |
| `esc` | back |
| `q` | quit (confirm dialog) |
| `Ctrl+T` | cycle theme |
| `Ctrl+C` | force quit |

## Architecture

| Layer | Choice |
|-------|--------|
| Language | Go 1.22+ |
| TUI framework | [bubbletea](https://github.com/charmbracelet/bubbletea) |
| Components | [bubbles](https://github.com/charmbracelet/bubbles) — spinner, progress, stopwatch |
| Forms | [huh](https://github.com/charmbracelet/huh) — select, multi-select, confirm |
| Styles | [lipgloss](https://github.com/charmbracelet/lipgloss) |
| Config format | TOML |
| Encryption | [age](https://github.com/FiloSottile/age) (planned) |
| Packaging | goreleaser → brew tap + `install.sh` (planned) |

## Repo layout

```
lfg/
├── cmd/
│   ├── lfg/main.go             entry — flag parsing, tea program
│   └── snap/main.go            screen → ANSI text helper for screenshots
├── internal/
│   ├── preset/                 bundle + tool data (hardcoded for prototype)
│   └── tui/
│       ├── app.go              screen state machine + global hotkeys
│       ├── theme.go            3 palettes + huh theme builder
│       ├── layout.go           Frame(): outer chrome, centering
│       ├── title.go            hand-drawn block logo + gradient sweep
│       ├── welcome.go          animated hero + numbered actions
│       ├── tree_picker.go      collapsible bundle/tool tree
│       ├── confirm.go          stats row + huh.Confirm
│       ├── progress.go         spinner + gradient bar + log tail
│       ├── done.go             next-steps card
│       ├── backup.go           huh.Confirm (encrypt) → spinner → result
│       └── quit_confirm.go     huh.Confirm dialog (`q` from anywhere)
└── assets/                     screenshots
```

## Development

```sh
make build           # ./lfg binary
make run             # build + run
make test            # snapshot tests across 6 widths × 3 themes (~0.6s)
make snap-update     # regenerate testdata/*.golden after intentional UI change
make preview                     # page through every golden in less -R
make preview ARGS=welcome        # filter goldens by substring
make widths          # static PNGs per width (needs `freeze`)
make demo            # animated gif (needs `vhs`)
make docker          # build container image (Ubuntu + linuxbrew)
make docker-run      # launch TUI in container
```

Single test:

```sh
go test ./internal/tui -run TestSnapshot_Welcome/welcome_lfg_md_100x30
```

Snapshots use direct `View()` calls (not teatest byte streams) so the
goldens are clean ANSI-stripped text — diffable in any review tool.

## Roadmap

**v0.1 (MVP)** — current slice

- [x] TUI skeleton with all screens
- [x] Animated gradient logo, tactical bulletin chrome
- [x] 3 themes, swappable via `--theme` or `Ctrl+T`
- [x] Tree picker with bundle/tool selection
- [x] Quit-confirm dialog
- [ ] Real package installers (brew, apt, mise)
- [ ] Fetch `default.toml` from presets repo
- [ ] Binary-present + version detection
- [ ] Real `lfg backup` → tar + optional age encrypt
- [ ] goreleaser + `install.sh`

**v0.2** — sync

- GitHub device-flow auth
- Dedicated `lfg-config` repo per user
- `lfg sync` / `lfg apply` / `lfg diff`
- Dotfile restore (not just backup)

**v0.3** — SSH + macOS

- `lfg ssh list` (wishlist-powered)
- `lfg ssh add-device` (fleet pubkey push)
- macOS `defaults` + `systemsetup` wizard
- `lfg doctor`

**Deferred / v1+**

- Hosted paid sync tier
- Plugin system for user-authored install recipes
- Teams / shared configs
- Windows support

## Design principles

1. **Wrap, don't reinvent.** chezmoi, mise, brew, age already work.
   `lfg` is the opinionated interactive layer on top.
2. **Use Charm components.** This project lives in the Charm
   ecosystem; if `huh`/`bubbles` ships a primitive, use it instead of
   hand-rolling. See `CLAUDE.md` for the rule.
3. **Never silently touch user data.** Private SSH keys stay on-device
   by default. Secrets are encrypted client-side; the master key never
   leaves the user's control.
4. **Beautiful beats clever.** First five minutes are the product. If
   it doesn't look great, nobody runs step two.
5. **Ship-able MVP.** UX-first prototype before installers, installers
   before sync, sync before the kitchen sink.

## Contributing

Early days. Open an issue before a PR so we can align on scope.

## License

TBD — likely MIT.
