---
title: Contributing
description: Open an issue first, run snapshots, keep PRs scoped.
---

Early days. Open an issue before a PR so we can align on scope.

## Local dev loop

```sh
make build           # ./lfg
make run             # build + run
make test            # snapshots, ~0.6s
make snap-update     # regenerate goldens after intentional UI change
```

Single test:

```sh
go test ./internal/tui -run TestSnapshot_Welcome/welcome_lfg_md_100x30
```

## Docker — fresh box, no host noise

```sh
make docker-test     # one-shot: build + auto-run lfg --debug + bash on exit
make docker-bare     # INCLUDE_BREW=0 — bare ubuntu, exercises brew bootstrap
make docker-shell    # bash inside fresh container, lfg on PATH
```

Each `--rm` container is a fresh `~/.config/lfg/`. Mount a volume
if you want state to persist.

## House rules

- **Use Charm components** — `huh` for forms, `bubbles` for
  primitives, `lipgloss` for style. Don't hand-roll buttons,
  selects, or inputs. See `CLAUDE.md` for the full table.
- **Don't touch `internal/tui/testdata/` by hand** — always
  `make snap-update` so widths/themes stay in sync.
- **One feature per PR.** Mixing UI + installer + docs makes
  review noisy.
- **Goldens in the diff.** Reviewers can eyeball them; that's the
  whole point of checking them in.

## What to avoid

The full list lives in `CLAUDE.md` ("Things to avoid"). Highlights:

- Re-enabling huh `Group.Base` border without removing the outer
  Frame card.
- Per-line `Align(Center)` on multi-line content (figlet, log
  tail) — use `PlaceHorizontal`.
- Re-computing `canvasW` inline in a screen — call
  `CanvasW(width)` instead.
- `Text: lipgloss.NoColor{}` in palettes — fails under
  docker / ssh / tmux + alt-screen. Use `AdaptiveColor`.
- `exec bash -l` as a "reload" command when the PATH block lives
  in `.bashrc`. Use plain `exec bash`.
- Forwarding raw subprocess stdout to the TUI log tail without
  stripping ANSI escapes.
- Treating `Tool.PostInstall` failures as step failures.

## License

TBD — likely MIT.
