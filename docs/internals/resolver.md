---
title: Version resolver
description: How the picker shows live latest versions without blocking the UI.
---

`internal/preset/resolve.go::ResolveVersion(ctx, t)` hits the right
registry per backend:

| Backend | Endpoint |
|---|---|
| npm-distributed packages | `registry.npmjs.org/<pkg>/latest` |
| brew formulas | `formulae.brew.sh/api/formula/<name>.json` |
| node-lts | `nodejs.org/dist/index.json` (first LTS entry) |
| curl bash-script installs | skipped |
| opaque mise-only plugins | skipped |

The tree picker fans these out via `tea.Cmd` on `Init`, then falls
back to the parsed `PlannedVersion` (parsed from `@<ver>` in the
install command) while in flight. Per-tool timeout is 3 s — never
blocks tree render.

## Render order

1. Picker mounts → every tool shows its parsed `PlannedVersion`
   (`v?` if none).
2. Concurrent registry probes fire.
3. Each `tea.Msg` that returns updates that one row in place.
4. At 3 s, any in-flight probe is cancelled and the row keeps
   showing the parsed pin.

This means the picker is interactive immediately, and "live"
versions populate over the next ~1 s without any spinner UI.

## Why not cache

The picker is the only place that uses the live version. The data
goes stale fast (npm packages publish constantly), and a stale
cache would surface as "your tree picker said v1.2 but I just
installed v1.4" — confusing at best. Re-querying on every launch
keeps the displayed value true.

## Adding a new backend

Add a switch case in `ResolveVersion`. Inputs: `Tool` struct
(`Name`, `Source`, `Pkg`, etc). Output: `(string, error)`. Honour
the context deadline.
