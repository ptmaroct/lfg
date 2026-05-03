---
title: Doctor
description: Diagnose environment readiness before installing.
---

```sh
lfg doctor
```

Runs a battery of environment checks and prints a green / amber /
red status for each:

- **OS + arch** — supported (darwin/linux × amd64/arm64) or not.
- **Network** — can reach `formulae.brew.sh`, `registry.npmjs.org`,
  `nodejs.org`. Without these the live version resolver will fall
  back to parsed defaults.
- **PATH writable** — `~/.local/bin` exists or can be created.
- **Shell rc detected** — at least one of `.bashrc`, `.zshrc`, fish
  config exists (so the `# lfg-managed PATH` block has a home).
- **Bootstrap-able backends** — checks whether `brew`, `mise`, `node`
  are present; if not, flags them as needing bootstrap.
- **State directory** — `~/.config/lfg/` is writable.

Exits non-zero if any check fails, suitable for use in CI or
provisioning scripts:

```sh
lfg doctor && lfg apply barebones
```

## Output

Each check prints one line:

```
✓ OS supported (linux/amd64)
✓ Network: brew formulae API reachable
✓ Network: npm registry reachable
⚠ Network: nodejs.org dist index slow (1.8s)
✓ PATH writable (~/.local/bin)
✓ Shell rc: ~/.bashrc
✓ Shell rc: ~/.zshrc
✓ Backend: brew (v4.4.16)
✗ Backend: mise (not found — will bootstrap)
✓ State dir: ~/.config/lfg/
```

`✗` is only fatal for `OS supported` and `State dir`. Everything
else is informational — `lfg apply` and the TUI bootstrap missing
backends as needed.

## Verbose mode

```sh
lfg doctor --verbose
```

Adds the resolved binary path, version, and timing for each check.
Useful when the registry probes feel slow.
