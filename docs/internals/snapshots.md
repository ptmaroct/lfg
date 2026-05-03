---
title: Snapshot tests
description: How the TUI is tested deterministically across widths and themes.
---

```sh
make test            # all goldens, ~0.6s
make snap-update     # regenerate testdata/*.golden after intentional change
make preview                       # page through every golden in less -R
make preview ARGS=welcome          # filter by substring
make preview ARGS="bundles xl"     # multi-arg = AND filter
```

## Why not teatest's byte-stream

`teatest`'s byte capture returns the full ANSI escape stream
including alt-screen sequences — useless for diffing. The test
suite bypasses `tea.Program` entirely:

1. `New(theme)` → root `Model`
2. Drive via direct `m.Update(...)` calls with synthetic key messages
3. Call `m.View()` to capture the rendered string
4. Diff against `testdata/<name>.golden`

The `cmd` returned by `Update` is resolved exactly once (via `cmd()`)
to surface synchronous transitions. Async cmds (timers, ticks) don't
fire — fine for our case because every screen renders meaningfully
on first paint.

## Update flag

Prefer `-update` over `-update-snap` (the `flag.Lookup("update")`
lookup catches teatest's reserved flag).

```sh
go test ./internal/tui -update -run TestSnapshot_Welcome
```

## Coverage matrix

Goldens cover **6 widths × 4 themes** for every screen — xs / sm /
md / lg / xl / xxl × lfg / dracula / catppuccin / colorblind. Adding
a screen means adding a corresponding `TestSnapshot_<Name>` in
`screens_test.go` that loops over the matrix.

## Mock data, never the host

`tui.New(theme)` injects `mockProgressRunner` for the install step;
CLI startup passes the real `installer.Run`. **Don't conflate the
two paths** — wiring real subprocesses into snapshot tests would
make goldens non-deterministic and slow.

The bundle/tool data the picker sees in tests is a deterministic
mock list, not the host system. `Tool.Installed` and `Tool.Version`
are stubbed; live registry probes are no-ops.

## When goldens diverge

`make test` shows the diff. If intentional:

```sh
make snap-update
git add internal/tui/testdata/
git commit -m "snap: update goldens for <change>"
```

Reviewers can eyeball the golden diff in the PR — that's the whole
point of checking them in.
