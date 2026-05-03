---
title: MCP servers
description: Install Model Context Protocol servers and auto-register them with every MCP-aware harness on your machine.
---

The `mcp` bundle ships **Model Context Protocol** servers — small
processes (or remote endpoints) that extend any MCP-aware AI harness
with new capabilities. lfg installs the server *and* wires it into
every supported harness present on your `$PATH`.

The bundle sits **directly below `skills`** in the tree picker, gated
on the same `node-lts` requirement (npm-distributed servers are still
the most common shape).

## What lfg does

For each MCP server you select:

1. **Install** — for stdio servers, runs `npm install -g <package>`
   so the first launch doesn't pay the `npx` download cost. For remote
   (HTTP / SSE) servers, this step is a no-op.
2. **Prompt for credentials** — between confirm and progress, lfg
   shows a **credentials wizard**: one masked input per unique env var
   declared by the selected MCPs. Inputs are deduped (a token shared
   by two servers is asked once). Leave any field blank to skip — you
   can `export FOO=...` in your shell rc later; the done screen
   reminds you of every blank.
3. **Register** — for every harness in the entry's `target_harnesses`
   list that's actually on `$PATH`, lfg invokes the harness's own
   `mcp add` command (or merges into its config file when there is no
   CLI helper). Collected credentials are baked into the harness
   config: stdio entries get `--env KEY=VAL`, remote entries get
   `${VAR}` placeholders in URLs/headers expanded inline.
4. **Surface unset env vars** — any var the user left blank in the
   wizard is shown on the **done screen** as a copy-paste
   `export FOO=...` reminder.

## Supported harnesses

| Harness | Binary | How lfg registers | Storage |
|---|---|---|---|
| Claude Code | `claude` | `claude mcp add --scope user` | `~/.claude.json` |
| Codex | `codex` | `codex mcp add` | `~/.codex/config.toml` |
| Droid | `droid` | `droid mcp add` | `~/.factory/mcp.json` |
| opencode | `opencode` | merge into config JSON | `~/.config/opencode/opencode.json` |

Adding a new harness to lfg means dropping a `harnessAdapter` into
`internal/installer/mcp_register.go` and adding it to the `init()`
list — every existing MCP entry then auto-extends to the new harness
on the next run (`target_harnesses = ["all"]` is the recommended
default for built-in entries).

## Schema

MCP entries live under any bundle (the default preset puts them in
`mcp`). Every field is documented in the [Preset TOML
schema](/reference/preset-schema); the MCP-specific subset is:

```toml
[[bundles.tools]]
source       = "mcp"           # routes to the mcp installer
name         = "my-server"     # display name; "mcp-" prefix stripped on register

# --- one of: stdio (default) | http | sse ---
mcp_type     = "stdio"

# --- stdio fields ---
mcp_package  = "@scope/server-name"   # npm package to install globally
mcp_command  = ""                     # optional: override launch binary (default: npx)
mcp_args     = []                     # extra args appended after the package

# --- remote (http / sse) fields ---
mcp_url      = "https://mcp.example.com/sse"
mcp_headers  = { "Authorization" = "Bearer ..." }   # sent on every request

# --- common ---
target_harnesses = ["all"]            # or e.g. ["claude-code", "codex"]
env_vars         = ["MY_SECRET"]      # surfaced in info card + done screen
```

The `target_harnesses = ["all"]` shorthand expands at install time to
every harness lfg knows about, so a custom preset stays in sync as new
harnesses are added.

## Custom preset example

Add a private remote MCP and ship it to every installed harness:

```toml
[[bundles]]
id   = "mcp"
name = "mcp"
description = "Custom MCP servers"

[[bundles.tools]]
name             = "internal-billing"
source           = "mcp"
description      = "Read internal billing tool"
mcp_type         = "http"
mcp_url          = "https://mcp.internal.corp/billing"
mcp_headers      = { "Authorization" = "Bearer $BILLING_TOKEN" }
target_harnesses = ["all"]
env_vars         = ["BILLING_TOKEN"]
```

Run `lfg --config ./preset.toml apply` and the entry registers with
every harness on your machine; `BILLING_TOKEN` shows up on the done
screen as a reminder to export it.

## Troubleshooting

- **"skipping registration: claude not on PATH"** — the harness isn't
  installed. Install it (it's in the `dev-tools` bundle) and re-run
  `lfg apply`; only the missing registrations re-execute.
- **"warning: claude-code registration failed — exit 1"** — a harness
  CLI rejected the add. The literal command lfg ran is in the log
  tail; copy it, run by hand, and the error message tells you what to
  fix (auth, conflicting name, malformed URL).
- **Set the env var first, *then* the MCP server connects** — harnesses
  inherit env from the parent shell at launch, not from the registered
  config. After exporting in your rc, restart the harness.
