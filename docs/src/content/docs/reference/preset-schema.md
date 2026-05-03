---
title: Preset TOML schema
description: Field-by-field reference for custom presets.
---

A preset is a TOML file with one or more `[[bundles]]` tables, each
containing one or more `[[bundles.tools]]` tables.

## Bundle fields

```toml
[[bundles]]
id = "barebones"          # required, unique, machine-readable
name = "barebones"        # required, displayed in picker
description = "..."       # optional, shown in info card
default = true            # optional, pre-selects bundle in picker
requires = ["node-lts"]   # optional, gate bundle until tool is selected
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `id` | string | yes | unique across all bundles |
| `name` | string | yes | shown in picker |
| `description` | string | no | one-liner, shown in info card |
| `default` | bool | no | pre-select in picker |
| `requires` | `[]string` | no | tool names that must be selected first |

## Tool fields

```toml
  [[bundles.tools]]
  name = "git"                                  # required
  source = "brew"                               # required
  homepage = "https://git-scm.com"
  description = "Distributed version control."
  install_mac = "brew install git"
  install_linux = "sudo apt-get install -y git"
  mandatory = false
  post_install = ["git config --global init.defaultBranch main"]
  skill_url = "https://..."                     # only for source=skills
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `name` | string | yes | binary name, also used for detect probe |
| `source` | string | yes | one of `brew`, `apt`, `mise`, `npm`, `curl`, `skills`, `custom` |
| `homepage` | string | no | shown in info card |
| `description` | string | no | shown in info card |
| `install_mac` | string | one of | install command on macOS |
| `install_linux` | string | one of | install command on Linux |
| `mandatory` | bool | no | always-on, can't be unchecked |
| `post_install` | `[]string` | no | shell commands chained after main install |
| `skill_url` | string | yes if `source=skills` | URL passed to `npx skills add` |

At least one of `install_mac` / `install_linux` is required.
`post_install` failures surface as warnings — the main install still
counts as success.

## Source backends

| Source | What it does |
|---|---|
| `brew` | `brew install <name>` (macOS + linuxbrew on Linux) |
| `apt` | `sudo apt-get install -y <name>` (Linux) |
| `mise` | `mise use -g <plugin>@<ver>` for language runtimes |
| `npm` | `npm i -g <pkg>` |
| `curl` | arbitrary `curl … \| sh` install scripts |
| `skills` | `npx skills add <skill_url>` |
| `custom` | the `install_mac` / `install_linux` strings as-is |

`brew`, `mise`, and `node` are bootstrap-able — if any selected tool
needs them and they're not on PATH, `lfg` installs them first.

## Version pins

`@<ver>` in any install command is parsed and shown in the picker as
`PlannedVersion` while the live registry probe is in flight:

```toml
install_mac = "npm i -g typescript@5.4.0"
```

If the registry probe succeeds, the picker shows `→ vX.Y.Z` (latest).
If it fails or times out, it falls back to the parsed pin.

## Sample

See [`testdata/sample-preset.toml`](https://github.com/ptmaroct/lfg/blob/main/testdata/sample-preset.toml)
in the repo for a complete working example.
