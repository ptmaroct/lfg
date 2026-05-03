# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`lfg` is an open-source TUI bootstrap CLI for new dev machines.
**v0.3 is shipped**: installers (brew/apt/mise/npm/curl/skills) execute
live with PATH augmentation between steps + PostInstall hooks for
prereqs, detect pass probes binaries, async live-version resolver hits
npm/brew/nodejs registries, `lfg backup` produces tar/tar.age, `lfg
doctor` runs environment checks, `lfg update` self-updates from GitHub
releases. Four themes (lfg/dracula/catppuccin/colorblind) cycle live.
Snapshot tests (`make test`) and the `cmd/snap` helper still see
deterministic mock data — never the host system.

Cobra-based subcommand surface: `lfg`, `lfg apply`, `lfg backup`,
`lfg doctor`, `lfg version`, `lfg update`. Default `lfg` (no args)
launches the TUI. Theme persists in `~/.config/lfg/state.json`.

## UI rule — always use Charm components

This project lives in the Charm ecosystem and ships `bubbletea`, `huh`,
`bubbles`, `lipgloss` already. **Use them. Do not hand-roll widgets when
a Charm component exists.**

| Need | Use | Don't write |
|------|-----|-------------|
| yes/no decision | `huh.NewConfirm` | custom Y/N button row |
| pick one of N | `huh.NewSelect` | custom cursor list |
| pick many | `huh.NewMultiSelect` | custom checkbox list |
| text input | `huh.NewInput` / `bubbles/textinput` | custom prompt |
| password / secret | `huh.NewInput().EchoMode(huh.EchoPassword)` | masked rolled-by-hand |
| filterable list | `bubbles/list` | custom delegate from scratch |
| spinner | `bubbles/spinner` | dot animation by hand |
| progress bar | `bubbles/progress` | block-string interpolator |
| stopwatch / timer | `bubbles/stopwatch` / `bubbles/timer` | time.Since loop |
| scrolling log | `bubbles/viewport` | manual line-tail slice |
| key help footer | `bubbles/help` (or our `KeyHint`/`HintLine`) | from scratch |
| paginator | `bubbles/paginator` | counter widget |

Style/theming goes through `lipgloss` + `HuhTheme(palette)` — no raw
ANSI escapes, no inline ad-hoc color constants.

When the design needs something Charm doesn't ship (e.g. the welcome
hero, the tactical stats row, the tree picker rows), build it on top of
`lipgloss` styles — but the *interactive* primitives must be Charm.
This keeps keybindings, theming, focus management, and a11y consistent
across the app.

If you find yourself reinventing buttons, checkboxes, selects, or text
inputs: stop and use `huh`.

## Common commands

```sh
make build           # ./lfg binary
make run             # build + run with default theme
make test            # snapshot tests across 6 widths × 3 themes (~0.6s)
make snap-update     # regenerate testdata/*.golden after intentional UI change
make preview                       # page through every golden in less -R
make preview ARGS=welcome          # filter goldens by substring
make preview ARGS="bundles xl"     # multi-arg = AND filter
make widths          # static PNGs per width (needs `brew install charmbracelet/tap/freeze`)
make demo            # record full flow as gif (needs `brew install vhs`)
make docker          # build container image (Ubuntu + linuxbrew)
make docker-run      # launch TUI in container
make docker-shell    # bash inside fresh container, lfg on PATH
make docker-test     # one-shot: build + auto-run lfg --debug + bash on exit
make docker-bare     # INCLUDE_BREW=0 — bare ubuntu, exercises brew bootstrap
make hooks           # one-time: wire .githooks/ for local commit-msg lint
```

Single test:
```sh
go test ./internal/tui -run TestSnapshot_Welcome/welcome_lfg_md_100x30
```

Module path matches the GitHub home: `github.com/ptmaroct/lfg`.

## Architecture

### Screen state machine (`internal/tui/app.go`)

`Model` holds a `screen` enum + a child model per screen. The root
`Update` dispatches messages to the active child; child models request
transitions by returning a `transitionMsg` cmd. The `transition` method
re-instantiates the destination child model with current selection state
so each screen always sees fresh, correct data.

**Selection state** flows screen → screen via two maps on the root model:
- `selectedBundleIDs map[string]bool` (set in bundle picker)
- `selectedTools map[string]bool` keyed `<bundleID>/<toolName>`

When adding a screen: add to the `screen` enum, add a child field on
`Model`, register dispatch in `Update`/`forwardSize`/`View`/`transition`,
and emit `goTo(screenX)` from wherever you trigger the transition.

### Theming (`internal/tui/theme.go`)

`Palette` is the abstract palette every screen uses (Primary, Accent,
Success, Muted, Gradient, etc.). Four palettes ship: `lfg`, `dracula`,
`catppuccin`, `colorblind` (IBM blue/orange/magenta). `HuhTheme(p)`
derives a `*huh.Theme` from a palette by cloning `huh.ThemeCharm()` and
overriding select fields. `Palette.Name` is read by `Frame` to show the
active theme name in the chrome top-right.

**The huh `Group.Base` border is intentionally hidden** — the visible
"vertical bars on the left" bug came from huh's group border showing
inside our outer Frame. Don't re-enable that border without removing
the outer Frame card too.

**Body Text uses AdaptiveColor, not NoColor.** Earlier iterations set
`Text: lipgloss.NoColor{}` to inherit terminal default fg. Inside
docker / over ssh / under tmux + alt-screen, that lookup misfires and
renders as black even on a dark bg — body text disappears entirely.
Pin Text to `lipgloss.AdaptiveColor{Light: ..., Dark: ...}` since
`HasDarkBackground` is explicitly set by `cli/root.go::applyBg`.

**Accent colors target Tailwind -400 stops** for the lfg theme
(pink-400 / violet-400 / emerald-400). The earlier -500 / -600
versions scored only ~3:1 against near-black terminals — small "01"
digits and `▸` cursor glyphs at body weight came out near-invisible.
-400 lands at ~5:1 against black while still readable on white.

Adding a theme: extend `PaletteFor` with a new case + add the name to
the `flag` validation in `cmd/cli/root.go` (search `validateTheme`) +
add to `screens_test.go`'s welcome theme loop so it gets snapshot
coverage.

### Layout (`internal/tui/layout.go`)

`Frame(palette, w, h, subtitle, inner, footer, compactTitle)` is the
single source of chrome. Every screen's `View(w, h)` ends in a `Frame`
call. Frame builds a fully closed bulletin box: `┏━━━┓` top with
corner glyphs, `┃` left + right edge per body line, `┗━━━┛` bottom.

**`CanvasW(width)` is the single source of truth for inner canvas
width.** Every screen calls it (don't recompute the formula inline);
having Frame and each screen do the math independently silently
desynced widths between transitions, and even a one-character typo
caused jitter. Bounds: 56 min, 100 max.

**Padding contract:** every body line is padded to `canvasW - 2`
BEFORE the vertical edges are added. Without uniform interior width,
short lines re-center inside their own bounding box and drift visibly
away from the rules. `padLinesTo` is what makes the box stay square.

`compactTitle` is `height < 22` — falls back from figlet banner to a
single-line gradient `lfg` string. Don't break this contract; xs/sm
test widths depend on it.

### Title (`internal/tui/title.go`)

Renders an ASCII figlet via `go-figure` then per-char gradient via the
hand-rolled `blend1D`. `lipgloss` v1.1.0 doesn't have built-in 1D color
blending; v2 does. If we upgrade lipgloss to v2, replace `blend1D` with
the upstream call.

**Font is `standard`** — switching fonts is a one-line change in
`title.go:25`. Bundled options: `standard`, `slant`, `big`, `block`,
`doom`, `larry3d`, `roman`, `script`, `shadow`, `term`, etc. `slant`
made `lfg` look slashy; `big` had `g` descender bleed; `standard` is
the cleanest balance for short brand names.

Every line is right-padded to `maxW` so downstream `Align(Center)` /
`PlaceHorizontal` treats the figlet block as one rectangle. Removing
that padding will break centering on the next render pass.

### Snapshot testing (`internal/tui/screens_test.go`)

**Don't use teatest's byte-stream capture for snapshots** — it returns
the full ANSI escape stream including alt-screen sequences, useless
for diffing. Instead, the test bypasses `tea.Program` entirely:

1. `New(theme)` → root Model
2. Drive via direct `m.Update(...)` calls with synthetic key messages
3. Call `m.View()` to capture the rendered string
4. Diff against `testdata/<name>.golden`

The `cmd` returned by `Update` is resolved exactly once (via `cmd()`)
to surface synchronous transitions. Async cmds (timers, ticks) don't
fire — fine for our case because every screen renders meaningfully on
first paint.

Update flag: prefer `-update` (the `flag.Lookup("update")` lookup
catches teatest's reserved flag). `-update-snap` is an alias.

When you change UI: `make test` shows diff. If intentional,
`make snap-update`, then commit `internal/tui/testdata/*.golden` so
reviewers can eyeball the diff in the PR.

### Preset data (`internal/preset/preset.go`)

Bundles + tools are hardcoded as the default; users can override via
`--config` (local TOML or http(s) URL). `Tool.Installed` + `Tool.Version`
get populated by the real detect pass at TUI startup. `Tool.PostInstall`
([]string) chains shell commands after the main install — used by
agent-browser to run `npm i -g agent-browser` + `agent-browser install`
after the skill stub lands.

### Version resolver (`internal/preset/resolve.go`)

`ResolveVersion(ctx, t)` hits the right registry per backend:
- npm-distributed packages → `registry.npmjs.org/<pkg>/latest`
- brew formulas → `formulae.brew.sh/api/formula/<name>.json`
- node-lts → `nodejs.org/dist/index.json` (first LTS entry)
- skipped: curl bash-script installs, opaque mise-only plugins

Tree picker fans these out via `tea.Cmd` on Init, falls back to the
parsed `PlannedVersion` (parsed from `@<ver>` in install commands)
while in flight. 3s per-tool timeout — never blocks tree render.

### Installer (`internal/installer/`)

**PATH augmentation between steps.** `installer.Run` prepends known
dev-bin dirs (`~/.local/bin`, mise shims, brew dirs) to `PATH` at
start, restores on exit, and re-augments before + after every step.
Without this, a `mise` install in step N landed in `~/.local/bin` but
step N+1's `sh -c` didn't see it, cascading "mise: not found" /
"npm: not found" / "npx: not found" exits across the rest of the
queue.

**Available pre-check + failedBackends.** Non-bootstrap steps call
`i.Available()` first; if false, emit a `skipped — <backend> not on
PATH` line and mark the backend failed. Subsequent same-backend steps
short-circuit with the same skip message instead of repeating exit-127.

**ANSI escape stripping in `scanLines`.** `npx skills add` (and
similar fancy installers) emit cursor-move + clear-line escapes inline
with stdout. Forwarding them to the TUI log tail scrambles cursor
positioning and breaks the bottom of the frame. `scanLines` regex-
strips `\x1b[...]` before storing each line.

**Plan dedup.** When a tool's name matches a bootstrap-able backend
(e.g. user explicitly selects `mise` AND a tool with `Source: "mise"`),
the auto-bootstrap is skipped — installing mise twice in one run was a
real bug.

**Live harness re-probe at skill-install time.** `skillsInstaller`
re-probes which AI harnesses are on PATH every time (not the cached
list from TUI startup). Reason: when the user installs codex + skills
in the same run, codex isn't on PATH at detect time and skills only
land in `~/.claude/skills/`. Live re-probe picks up codex installed
earlier in the run.

**PostInstall failures = warnings, not step failures.** The main
install (skill stub / brew install / etc) already succeeded; flagging
the whole step for a chrome-on-arm64 mismatch was misleading. Surface
a `(main install still succeeded)` line and continue.

### Shell rc augmentation (`internal/installer/shellrc.go`)

`EnsureShellPath()` writes a fenced `# lfg-managed PATH` block to
EVERY detected rc (`~/.bashrc`, `~/.zshrc`, fish config), idempotent
per file. Multi-shell so a future `chsh` doesn't lose the
augmentation. Idempotency hinges on the exact marker strings — never
change them without a migration path that prunes old blocks first.

**Always target the INTERACTIVE rc** (`.bashrc`, not `.bash_profile`)
so the matching `exec bash` reload command actually picks up the new
PATH — login-shell bash on Linux sources `.bash_profile`, not
`.bashrc`. Done screen's `reloadShellCmd` deliberately drops the `-l`
flag for that reason.

**Shell detection priority:**
1. `$SHELL` env var
2. `/proc/<ppid>/comm` parent process name (covers `docker run -it
   bash` where `$SHELL` stays unset and the fallback used to wrongly
   write to `~/.zshrc`)
3. Existing-file sniff
4. Last resort: linux→.bashrc, darwin→.zshrc

## Docker

`Dockerfile` is multi-stage: Go build → Ubuntu 24.04 + (optional)
Homebrew. The brew install runs as non-root user `dev` (brew refuses
root on Linux). Image ~1.05GB with brew, ~250MB without.

**`INCLUDE_BREW` build-arg** toggles brew preinstall. Default `1`
(fast iteration; brew shows under ALREADY INSTALLED). `0` produces a
bare ubuntu image to exercise lfg's full brew bootstrap path.

**Cache placement matters.** `ARG INCLUDE_BREW` is declared
immediately before the conditional brew RUN — NOT at the top of the
stage. BuildKit folds every in-scope ARG value into the cache key of
every subsequent RUN (even ones that don't reference the ARG); putting
ARG at the top of the stage meant `docker-bare` (INCLUDE_BREW=0) and
`docker` (INCLUDE_BREW=1) had different cache keys for the expensive
apt-get + useradd layers, forcing 2-min full rebuilds on every variant
switch. Declaring it later lets the heavy ubuntu setup share cache
across both variants.

`make docker-test` is the tightest iteration loop: rebuild + auto-run
`lfg --debug` + drop into bash on exit. Source edits trigger only the
Go-build + COPY layers (~10s warm cache).

## Plan + roadmap

The agent plan that drove the design lives outside the repo at
`/Users/anuj/.claude/plans/what-are-some-of-eventual-coral.md` (also
copied to `plan.md` which is gitignored). README has the v0.1 → v1
roadmap. v0.2 = GitHub auth + remote sync. v0.3 = SSH + macOS defaults.
Anything not in v0.1 should not land without revisiting that plan.

## Commit messages — Conventional Commits, enforced

**All commits and PR titles must follow Conventional Commits.**
The `.github/workflows/commitlint.yml` workflow blocks PR merges if
the title or any commit on the branch fails. The local
`.githooks/commit-msg` hook (wired via `make hooks`) catches the same
failures pre-push.

**Format:** `<type>(<optional-scope>)!?: <subject>`

Subject must:
- start with a **lowercase** letter
- be ≥4 chars and ≤100 chars
- **not end with a period**
- describe the change in the imperative ("add", not "added"/"adds")

**Allowed types** (mirrors `commitlint.config.mjs`):
`feat`, `fix`, `perf`, `refactor`, `docs`, `test`, `build`, `ci`,
`chore`, `style`, `revert`, `release`.

**Release-impact mapping** (release-please derives the next version):
- `feat:` → minor bump
- `fix:` / `perf:` → patch bump
- `feat!:` or `BREAKING CHANGE:` footer → major bump (or minor while
  pre-1.0 because `bump-minor-pre-major` is set)
- `docs:` / `chore:` / `test:` / `ci:` / `style:` / `build:` /
  `refactor:` / `release:` → no release triggered

**Good examples:**
```
feat(installer): add support for fish shell rc
fix: respect XDG_CONFIG_HOME on Linux
docs(readme): document brew install flow
ci: bump goreleaser-action to v6
feat!: drop legacy v0.1 preset format
```

**Common ways to fail the hook:**
- `Fix bug` → must be `fix: <something>` (lowercase type, colon, scope)
- `feat: Added new theme` → subject starts uppercase + past tense
- `feat: add theme.` → trailing period
- `update README` → no type prefix

**For squash merges** (which is how this repo merges PRs), the **PR
title** becomes the commit message on `main` — release-please reads
that, not the inner commits. The `pr-title` job in commitlint.yml
enforces the same rules on PR titles. So when opening a PR, write the
title as if it's the final commit message.

**Co-author trailers** (the `Co-Authored-By:` line) and other footers
are always allowed; the rules apply only to the header line.

## Release pipeline

**Two distribution channels, one tap.** `ptmaroct/homebrew-tap`
holds two formulae:

- `Formula/lfg.rb`     → stable;  `brew install ptmaroct/tap/lfg`
- `Formula/lfg-beta.rb` → preview; `brew install ptmaroct/tap/lfg-beta`

**Branches drive channels.**

- Push to `main` → `release-please.yml` opens/updates a "Release
  vX.Y.Z" PR with conventional-commit changelog + manifest bump.
  Merging the PR tags `vX.Y.Z` → `release.yml` runs goreleaser →
  binaries land on GitHub Releases AND `lfg.rb` is rewritten in the
  tap.
- Push to `develop` → `release-please-beta.yml` opens "Release
  vX.Y.Z-beta.N" PR. Merge tags the prerelease → goreleaser publishes
  prerelease binaries AND rewrites `lfg-beta.rb` only (the stable
  `lfg.rb` is untouched because of `skip_upload: auto` on prereleases).

**Why two release-please workflows + configs?** Each channel needs
its own `manifest` file (`.release-please-manifest.json` vs
`.release-please-manifest-beta.json`) so release-please can track the
last-released version per channel independently. Sharing one manifest
would make beta increments overwrite stable's last-released marker.

**release-please vs goreleaser ownership.** Both workflows set
`skip-github-release: true` on the release-please-action input. That's
deliberate — without it, release-please creates a draft GitHub Release
the moment it tags, then goreleaser tries to create another release
with the same tag and fails (`release already exists`). With
skip-github-release, release-please owns the PR + CHANGELOG.md + tag;
goreleaser owns the GitHub Release + artifacts + tap formula push.

**Goreleaser two-formula split.** `.goreleaser.yaml` has two `brews:`
entries pointing at the same tap repo. The `lfg` entry uses
`skip_upload: auto` so it's automatically skipped on snapshots and
prereleases. The `lfg-beta` entry uses an inverse template
(`{{ if and (not .IsSnapshot) .Prerelease }}false{{ else }}true{{ end }}`)
to publish only on prereleases.

**HOMEBREW_TAP_TOKEN.** Repo secret on `ptmaroct/lfg`; fine-grained
PAT scoped to `ptmaroct/homebrew-tap` with `Contents: read+write`. Set
via `printf '%s' '<token>' | gh secret set HOMEBREW_TAP_TOKEN --repo
ptmaroct/lfg`. **Do NOT use `--body -`** — gh treats the dash as a
literal value, not a stdin marker, and you'll silently set the secret
to the string "-". Prior outage took two CI runs to diagnose because
the goreleaser logs show "***" for both a real token and the literal
"-".

**Manual fallback for a stuck release.** If the formula push fails
but binaries published, do not hand-craft a `Formula/lfg.rb` and push
it to the tap as if goreleaser wrote it — the impersonation breaks
trust + bypasses signing. Instead: fix the secret, then bump to the
next patch version (e.g. v0.3.1 broken → tag v0.3.2). Re-using the
same tag means new tarballs with different BuildDates → different
SHA256s, which would break anyone who downloaded between attempts.

**Goreleaser deprecation warning.** `brews:` shows a "phasing out in
favor of homebrew_casks" warning at run time. Misleading — `brews:`
is for formulae, `homebrew_casks:` is for casks (gui apps). The
warning can be ignored until goreleaser actually breaks the key. If
they ever do, both formula entries need migrating.

## Things to avoid

- Re-enabling huh `Group.Base` border without removing the outer Frame card.
- Per-line `Align(Center)` on multi-line content (figlet, log tail) — use
  `PlaceHorizontal` so blocks center as rectangles.
- Touching `internal/tui/testdata/` by hand — always regenerate via
  `make snap-update` so all themes/widths stay in sync.
- Wiring real subprocesses into snapshot tests — `tui.New(theme)` uses
  `mockProgressRunner`; CLI startup passes `installer.Run`. Don't
  conflate the two paths.
- Re-computing `canvasW` inline in a screen — call `CanvasW(width)`.
  Drift between Frame and screens silently shifts widths between
  transitions.
- `ARG` declarations at the top of a Dockerfile stage when only a
  later RUN needs them — taints all upstream layer cache keys.
- `Text: lipgloss.NoColor{}` in palettes — fails to render under
  docker / ssh / tmux + alt-screen. Use `AdaptiveColor`.
- `exec bash -l` as a "reload" command when the PATH block lives in
  `.bashrc`. Login-shell bash sources `.bash_profile` instead. Plain
  `exec bash` re-runs as interactive (since stdin is a tty) and
  sources the right file.
- `gh secret set NAME --body -` thinking the dash is a stdin marker —
  it's literal, and you'll set the secret to "-". Use plain
  `printf '%s' "$value" | gh secret set NAME` instead.
- Hand-writing `Formula/lfg.rb` and pushing it to the tap to "unblock"
  a failed release. Goreleaser's signing/authorship is the trust
  anchor; manual pushes bypass it. See "Release pipeline" above.
- Re-using a tag (e.g. re-pushing v0.3.1 after deleting it) when
  binaries were briefly public. New goreleaser run produces new
  BuildDate-baked tarballs with different SHA256s; anyone who pulled
  the first set is broken. Always bump to the next patch instead.
- Letting both release-please and goreleaser create the GitHub Release.
  Set `skip-github-release: true` on the release-please-action input;
  goreleaser is the sole release publisher.
- Forwarding raw subprocess stdout to the TUI log tail without
  stripping ANSI escapes — fancy installers (npx skills add) emit
  cursor-move sequences that scramble the frame.
- Treating `Tool.PostInstall` failures as step failures. Main install
  already succeeded; surface the failure as a warning and continue.

## Cross-platform gotchas

- macOS Screenshot filenames use NNBSP (U+202F) before `PM`/`AM`, not
  a regular space. Quoted regular-space paths fail; use a glob
  (`Screenshot*9.38.24*PM.png`) or read the raw bytes.
