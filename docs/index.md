---
layout: home

hero:
  name: lfg
  text: Make a new dev machine feel like home.
  tagline: Open-source TUI bootstrap CLI — minutes, not hours.
  image:
    src: /logo.svg
    alt: lfg
  actions:
    - theme: brand
      text: Install
      link: /install
    - theme: alt
      text: Quick start
      link: /quick-start
    - theme: alt
      text: View on GitHub
      link: https://github.com/ptmaroct/lfg

features:
  - title: Tree picker
    details: Bundles + tools as a collapsible tree. Live versions resolved from npm / brew / nodejs registries.
  - title: Real installers
    details: brew · apt · mise · npm · curl · skills. PATH augmented between steps so each step sees what the previous one installed.
  - title: Backup & restore
    details: "`lfg backup` snapshots dotfiles + configs into tar.gz or tar.age (encrypted with age)."
  - title: Themes
    details: Four built-in palettes (lfg / dracula / catppuccin / colorblind). Cycle live with Ctrl+T.
---

## Why

Setting up a fresh machine eats a full evening: install Homebrew, mise,
node, bun, go, …; pull dotfiles; copy SSH config; tune a dozen macOS
preferences; miss two and notice them next Tuesday.

`lfg` wraps the best-in-class primitives — Homebrew, mise, chezmoi, age
— in a [Charm](https://charm.sh) TUI so the first five minutes on a new
box feel great.

![welcome](/screens/welcome.png)
