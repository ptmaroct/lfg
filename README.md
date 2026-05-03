# lfg

> Make a new dev machine feel like home in minutes.

`lfg` is an open-source TUI that bootstraps a fresh Mac or Linux dev
box — installs the tools you pick, restores your dotfiles, and (soon)
distributes your SSH identity to the servers you trust.

📖 **Full docs:** [lfg-docs.netlify.app](https://lfg-docs.netlify.app)

![welcome](assets/welcome.png)

## Install

Homebrew (macOS + Linux):

```sh
brew install ptmaroct/tap/lfg
```

Want the preview channel? `brew install ptmaroct/tap/lfg-beta` follows the
`develop` branch — newer features, fewer guarantees.

Or via the curl one-liner:

```sh
curl -fsSL https://raw.githubusercontent.com/ptmaroct/lfg/main/install.sh | sh
```

Or from source (Go 1.25+):

```sh
git clone https://github.com/ptmaroct/lfg && cd lfg && make build && ./lfg
```

## Quick start

```sh
lfg              # launch the TUI
lfg --help       # all subcommands and flags
```

## The journey

### 1. Land

Drop into the welcome screen with a quick rundown of what `lfg` will
do on this box.

![welcome](assets/welcome.png)

### 2. Pick

Bundles + tools as a collapsible tree. Already-installed tools are
flagged with their current version so you don't reinstall.

![pick](assets/tree.png)

### 3. Confirm

See the exact `brew install` / `apt-get install` / `mise use` /
`npm i -g` commands that will run, before they run.

![confirm](assets/confirm.png)

### 4. Install

Each step streams live. PATH is re-augmented between steps so the
next installer always sees what the previous one put down.

![install](assets/install.png)

## Why

Setting up a fresh machine eats an evening: brew, mise, node, dotfiles,
SSH config, macOS preferences. `lfg` wraps the best-in-class primitives
(Homebrew, mise, chezmoi, age) in a [Charm](https://charm.sh) TUI so
the first five minutes on a new box feel great.

**Status:** v0.3 — installers, PostInstall hooks, async live version
resolver, `lfg backup`, `lfg doctor`, `lfg update`, four themes.

## Develop

```sh
make build           # ./lfg
make test            # snapshot tests across widths × themes
make snap-update     # regenerate goldens after intentional UI change
make docker-test     # one-shot: build + auto-run lfg --debug + bash on exit
make docs            # local docs site (http://localhost:5173/)
```

See [docs/project/contributing](https://lfg-docs.netlify.app/project/contributing)
for the full list and house rules.

## Contributing

Open an issue before a PR so we can align on scope. Roadmap:
[docs/project/roadmap](https://lfg-docs.netlify.app/project/roadmap).

## License

MIT.

---

Built by **[Anuj Sharma](https://anujsh.com)** · [@waahbete](https://x.com/waahbete) · [GitHub](https://github.com/ptmaroct/lfg)
