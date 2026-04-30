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

## Demo

```
$ lfg

╭─────────────────── lfg ───────────────────╮
│                                            │
│              [gradient banner]             │
│         bootstrap your dev machine         │
│                                            │
│   ╭─ What would you like to do? ───╮       │
│   │  ▸ Install recommended setup   │       │
│   │    Backup this machine to file │       │
│   │    Quit                        │       │
│   ╰────────────────────────────────╯       │
│                                            │
│   ↑/↓ move · enter select · ctrl-c quit    │
╰────────────────────────────────────────────╯
```

Flow:

1. **Welcome** — pick install or backup.
2. **Bundles** — stack curated sets: `default`, `mac-power-user`,
   `ai-clis`, `media`.
3. **Tools** — fuzzy-filterable multi-select, shows installed version
   vs. "not installed" for each.
4. **Confirm** — summary by source (brew, apt, mise, custom).
5. **Progress** — gradient progress bar, live log, elapsed timer.
6. **Done** — next-steps card.

Backup flow: confirm encrypt yes/no → write `.tar[.age]` → print path
plus key-backup reminder.

## Install (prototype)

Requires Go 1.22+.

```sh
git clone https://github.com/anuj/lfg
cd lfg
go build -o lfg ./cmd/lfg
./lfg
```

Once the MVP lands, this will become:

```sh
curl -fsSL https://raw.githubusercontent.com/lfg-cli/lfg/main/install.sh | sh
```

## Usage

```sh
./lfg                      # launch TUI
./lfg --theme=lfg          # default: pink → purple → mint
./lfg --theme=dracula      # classic Dracula
./lfg --theme=catppuccin   # Catppuccin Mocha
```

## Architecture

| Layer | Choice |
|-------|--------|
| Language | Go 1.22+ |
| TUI framework | [bubbletea](https://github.com/charmbracelet/bubbletea) |
| Components | [bubbles](https://github.com/charmbracelet/bubbles) — spinner, progress, stopwatch |
| Forms | [huh](https://github.com/charmbracelet/huh) — select, multi-select, confirm |
| Styles | [lipgloss](https://github.com/charmbracelet/lipgloss) |
| ASCII banner | [go-figure](https://github.com/common-nighthawk/go-figure) |
| Config format | TOML |
| Encryption | [age](https://github.com/FiloSottile/age) (planned) |
| Packaging | goreleaser → brew tap + `install.sh` (planned) |

## Repo layout

```
lfg/
├── cmd/lfg/main.go              entry — flag parsing, tea program
└── internal/
    ├── preset/                  bundle + tool data (hardcoded for prototype)
    └── tui/
        ├── app.go               screen state machine
        ├── theme.go             3 palettes + huh theme builder
        ├── layout.go            Frame(): outer chrome, centering
        ├── title.go             gradient figlet banner
        ├── welcome.go           huh.Select screen
        ├── bundle_picker.go     huh.MultiSelect for bundle stacking
        ├── tool_picker.go       huh.MultiSelect per-bundle, filterable
        ├── confirm.go           huh.Confirm + summary panel
        ├── progress.go          spinner + gradient bar + log tail
        ├── done.go              next-steps card
        └── backup.go            encrypt y/n → pack → result card
```

## Key bindings

- `↑ ↓ j k` — move
- `space x` — toggle option
- `/` — filter (tool picker)
- `tab shift-tab` — switch between bundles (tool picker)
- `enter` — confirm / continue
- `esc` — back
- `ctrl-c` — quit

## Roadmap

**v0.1 (MVP)** — current slice

- [x] TUI skeleton with all screens
- [x] 3 themes, swappable via `--theme`
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
2. **Never silently touch user data.** Private SSH keys stay on-device
   by default. Secrets are encrypted client-side; the master key never
   leaves the user's control.
3. **Beautiful beats clever.** First five minutes are the product. If
   it doesn't look great, nobody runs step two.
4. **Ship-able MVP.** UX-first prototype before installers, installers
   before sync, sync before the kitchen sink.

## Contributing

Early days — the plan is at
[`../.claude/plans/what-are-some-of-eventual-coral.md`](../.claude/plans/what-are-some-of-eventual-coral.md)
(will move into this repo). Open an issue before a PR so we can align
on scope.

## License

TBD — likely MIT.
