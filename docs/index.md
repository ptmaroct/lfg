---
layout: home

hero:
  text: Feel back at home, in minutes.
  tagline: "lfg is the day-one CLI for developers. Pick the tools, restore the dotfiles, walk away — come back to a machine that already feels yours."
  actions:
    - theme: brand
      text: Install in 30s
      link: /install
    - theme: alt
      text: Quick start
      link: /quick-start
---

![welcome](/screens/welcome.png)

## Get started

```sh
curl -fsSL https://raw.githubusercontent.com/ptmaroct/lfg/main/install.sh | sh
lfg
```

That's it. Pick what you want. `lfg` installs it.
[Full install guide →](/install)

## How it works

1. **Pick.** Bundles + tools land in a collapsible tree. Anything already
   installed is flagged with its current version so you don't reinstall it.
2. **Confirm.** See the exact commands that will run before they run —
   `brew install`, `apt-get install`, `mise use`, `npm i -g`, the lot.
3. **Watch.** Each step streams live. The shell PATH is re-augmented
   between steps so the next installer always sees the previous one.
4. **Back up.** `lfg backup` snapshots dotfiles + configs into a tarball
   (encrypted, optionally) — so the *next* day-one is shorter than this one.

## Why it exists

You know the drill. New laptop arrives — fast, blank, slightly hostile.
An hour of `brew install`. Then mise, then node, then bun, then go. Hunt
down dotfiles. Copy SSH config across. Tweak macOS preferences and forget
two of them. By 7pm you've barely written a line of code.

`lfg` makes a new box feel like the old one. Pick the tools, hit go,
walk away. When you're done, `lfg backup` snapshots everything for the
next machine.

## Roadmap

- **v0.4** — GitHub auth + remote sync (back up to your own repo)
- **v0.5** — SSH identity distribution to trusted servers
- **v0.6** — macOS defaults pane

[Full roadmap →](/project/roadmap)
