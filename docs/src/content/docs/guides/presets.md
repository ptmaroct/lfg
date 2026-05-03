---
title: Custom presets
description: Define your own bundles and tools in TOML.
---

The built-in bundles are a good default but probably not *yours*.
Pass any TOML file as `--config` to swap in your own:

```sh
lfg --config ./my-preset.toml
lfg --config https://example.com/preset.toml   # http(s) supported
```

## Minimal example

```toml
[[bundles]]
id = "minimal"
name = "minimal"
default = true

  [[bundles.tools]]
  name = "git"
  source = "brew"
  homepage = "https://git-scm.com"
  install_mac = "brew install git"
  install_linux = "sudo apt-get install -y git"
```

`default = true` pre-selects the bundle in the picker. Mark a tool
`mandatory = true` to make it always-on (rendered with `[●]`, can't
be unchecked).

## Round-trip your current setup

```sh
lfg export                   # save → ~/lfg-preset-YYYY-MM-DD.toml
lfg export -o ./preset.toml  # explicit path
lfg --config ./preset.toml   # load it (bundles, mandatories, install commands all preserved)
```

## Field reference

See [Preset TOML schema](/lfg/reference/preset-schema/) for the full
field-by-field reference (post-install hooks, skills URLs, source
backends, etc).

## Validation

`lfg --config` parses the file before launching the TUI. Common
errors:

- **unknown source** — must be one of `brew`, `apt`, `mise`, `npm`,
  `curl`, `skills`, `custom`.
- **missing install command** — at least one of `install_mac` /
  `install_linux` is required.
- **skills without skill_url** — `source = "skills"` needs
  `skill_url = "https://..."` so `npx skills add` knows what to fetch.

`lfg doctor` (without `--config`) prints validation hints for the
default preset.
