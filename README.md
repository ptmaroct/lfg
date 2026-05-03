# lfg

> Make a new dev machine feel like home in minutes.

`lfg` is an open-source TUI that bootstraps a fresh Mac or Linux dev
box — installs the tools you pick, restores your dotfiles, and (soon)
distributes your SSH identity to the servers you trust.

📖 **Full docs:** <https://lfg-docs.netlify.app>

![welcome](assets/welcome.png)

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/ptmaroct/lfg/main/install.sh | sh
```

Or from source (Go 1.25+):

```sh
git clone https://github.com/ptmaroct/lfg && cd lfg && make build && ./lfg
```

## Quick start

```sh
lfg                              # launch TUI
lfg apply barebones              # headless install of one bundle
lfg backup                       # snapshot dotfiles + configs
lfg --config ./preset.toml       # custom bundles
lfg doctor                       # diagnose env readiness
lfg update                       # self-update from GitHub
```

`--dry-run` (or `-n`) walks any flow without exec'ing anything.
`--theme=dracula|catppuccin|colorblind|lfg` (or `Ctrl+T` to cycle).

## Screens

| Tree picker | Confirm | Install |
|---|---|---|
| ![tree](assets/tree.png) | ![confirm](assets/confirm.png) | ![install](assets/install.png) |

## Why

Setting up a fresh machine eats an evening: brew, mise, node, dotfiles,
SSH config, macOS preferences. `lfg` wraps the best-in-class primitives
(Homebrew, mise, chezmoi, age) in a [Charm](https://charm.sh) TUI so
the first five minutes on a new box feel great.

**Status:** v0.3 — installers, PostInstall hooks, async live version
resolver, `lfg backup`, `lfg doctor`, `lfg update`, four themes.

## Architecture (one-liner)

Go 1.25+ • bubbletea/bubbles/huh/lipgloss • cobra • TOML preset •
age-encrypted backups • goreleaser releases. Full breakdown:
[docs/internals/architecture](https://lfg-docs.netlify.app/internals/architecture/).

## Develop

```sh
make build           # ./lfg
make test            # snapshot tests across widths × themes
make snap-update     # regenerate goldens after intentional UI change
make docker-test     # one-shot: build + auto-run lfg --debug + bash on exit
make docs            # local docs site (http://localhost:4321/)
```

See [docs/project/contributing](https://lfg-docs.netlify.app/project/contributing/)
for the full list and house rules.

## Contributing

Open an issue before a PR so we can align on scope. Roadmap:
[docs/project/roadmap](https://lfg-docs.netlify.app/project/roadmap/).

## License

TBD — likely MIT.
