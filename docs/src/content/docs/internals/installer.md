---
title: Installer & PATH
description: How install steps run, why PATH is re-augmented every step, and how skipped steps work.
---

The installer (`internal/installer/`) wraps each backend in a
streaming `exec.Cmd` and drives them in sequence. Three behaviours
are worth calling out because they fix real bugs we hit.

## PATH augmentation between steps

`installer.Run` prepends known dev-bin dirs (`~/.local/bin`, mise
shims, brew dirs) to `PATH` at start, restores on exit, and
re-augments before + after every step.

Without this, a `mise` install in step N landed in `~/.local/bin` but
step N+1's `sh -c` didn't see it, cascading
`mise: not found` / `npm: not found` / `npx: not found` exits across
the rest of the queue.

## `Available()` pre-check + failedBackends

Non-bootstrap steps call `i.Available()` first; if false, emit a
`skipped — <backend> not on PATH` line and mark the backend failed.
Subsequent same-backend steps short-circuit with the same skip
message instead of repeating exit-127 noise.

## ANSI escape stripping

`scanLines` strips `\x1b[...]` escapes before forwarding output to
the TUI log tail. Some installers (`npx skills add`, certain brew
formulas) emit cursor-move + clear-line sequences inline with stdout;
forwarding them raw scrambles the bottom of the frame.

## Plan dedup

When a tool's name matches a bootstrap-able backend (e.g. user
explicitly selects `mise` AND a tool with `Source: "mise"`), the
auto-bootstrap is skipped — installing mise twice in one run was a
real bug.

## Live harness re-probe at skill-install time

`skillsInstaller` re-probes which AI harnesses are on PATH every
time, instead of using the cached list from TUI startup. Reason:
when a user installs `codex` + skills in the same run, codex isn't on
PATH at detect time and skills only land in `~/.claude/skills/`.
Live re-probe picks up codex installed earlier in the run.

## PostInstall failures = warnings

The main install (skill stub / brew install / etc) already
succeeded; flagging the whole step for a Chrome-on-arm64 mismatch in
`agent-browser install` was misleading. Surface a `(main install
still succeeded)` line and continue.

## Shell rc augmentation

`internal/installer/shellrc.go::EnsureShellPath()` writes a fenced
`# lfg-managed PATH` block to **every** detected rc (`~/.bashrc`,
`~/.zshrc`, fish config), idempotent per file. Multi-shell so a
future `chsh` doesn't lose the augmentation.

Always targets the **interactive** rc (`.bashrc`, not
`.bash_profile`) so the matching `exec bash` reload command actually
picks up the new PATH — login-shell bash on Linux sources
`.bash_profile`, not `.bashrc`.

Shell detection priority:

1. `$SHELL` env var
2. `/proc/<ppid>/comm` parent process name (covers `docker run -it
   bash` where `$SHELL` stays unset)
3. Existing-file sniff
4. Last resort: linux → `.bashrc`, darwin → `.zshrc`
