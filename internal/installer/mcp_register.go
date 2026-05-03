package installer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ptmaroct/lfg/internal/preset"
)

// harnessAdapter wires one MCP server entry into a specific AI-coding
// harness's config. Adapters MUST be idempotent — running them twice
// with the same tool should be a no-op (or a benign overwrite).
//
// Add a new harness:
//  1. Implement harnessAdapter for the new harness.
//  2. Register it in init() below.
//  3. Done — every existing MCP preset entry that uses the "all"
//     shorthand (or lists the new harness explicitly) starts wiring
//     into it on the next `lfg apply`.
type harnessAdapter interface {
	// Name is the canonical ID used in preset target_harnesses lists
	// (e.g. "claude-code", "codex", "opencode", "droid").
	Name() string
	// Binary is the CLI binary lfg checks via PATH lookup before
	// attempting registration. Missing binary → skip silently.
	Binary() string
	// Register wires the MCP server entry into this harness's config.
	// Implementations either shell out to a `<harness> mcp add` CLI
	// (preferred — handles merging + scope correctly) or merge into
	// the harness's config file directly when no CLI helper exists.
	Register(ctx context.Context, t preset.Tool, out chan<- Line) error
}

var harnessAdapters = map[string]harnessAdapter{}

func registerHarnessAdapter(a harnessAdapter) { harnessAdapters[a.Name()] = a }

// KnownHarnesses returns the sorted list of every registered harness.
// Used as the default target list when a preset entry omits
// `target_harnesses` or sets it to ["all"].
func KnownHarnesses() []string {
	out := make([]string, 0, len(harnessAdapters))
	for name := range harnessAdapters {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func init() {
	registerHarnessAdapter(claudeCodeAdapter{})
	registerHarnessAdapter(codexAdapter{})
	registerHarnessAdapter(droidAdapter{})
	registerHarnessAdapter(opencodeAdapter{})
}

// mcpServerName returns the registration name passed to harness CLIs.
// We strip a leading `mcp-` prefix so `mcp-github` registers as
// `github` — the harness namespace is already MCP-only, double-prefix
// reads ugly in `claude mcp list`.
func mcpServerName(t preset.Tool) string {
	return strings.TrimPrefix(t.Name, "mcp-")
}

// stdioLaunchParts returns the launch argv for a stdio MCP entry.
// MCPCommand wins when set; otherwise default to npx with -y so the
// package is auto-fetched if a fresh install is missing.
func stdioLaunchParts(t preset.Tool) []string {
	if t.MCPCommand != "" {
		return append([]string{t.MCPCommand}, t.MCPArgs...)
	}
	return append([]string{"npx", "-y", t.MCPPackage}, t.MCPArgs...)
}

// expandCreds replaces ${VAR} occurrences in s with values from the
// tool's MCPCredentials map. Unknown / blank vars expand to "" — the
// adapter then drops empty headers / URL fragments rather than writing
// a literal `${VAR}` into the harness config.
func expandCreds(s string, t preset.Tool) string {
	if len(t.MCPCredentials) == 0 || !strings.Contains(s, "${") {
		return s
	}
	for k, v := range t.MCPCredentials {
		s = strings.ReplaceAll(s, "${"+k+"}", v)
	}
	// Strip any unresolved ${...} placeholders so we don't write garbage.
	for {
		i := strings.Index(s, "${")
		if i < 0 {
			break
		}
		j := strings.Index(s[i:], "}")
		if j < 0 {
			break
		}
		s = s[:i] + s[i+j+1:]
	}
	return s
}

// hasCredentialValue returns true when t has a non-empty value for
// every var listed in t.EnvVars. When false, callers fall back to
// not passing --env (or skip the header), so the user can set the
// var in their shell rc and re-run.
func hasAllCreds(t preset.Tool) bool {
	for _, v := range t.EnvVars {
		if t.MCPCredentials[v] == "" {
			return false
		}
	}
	return true
}

// stdioEnvFlags renders --env KEY=VALUE pairs (in sorted key order so
// the rendered command is stable across runs) for each EnvVar that has
// a collected credential value. Empty values are skipped.
func stdioEnvFlags(t preset.Tool, flag string) string {
	if len(t.EnvVars) == 0 || len(t.MCPCredentials) == 0 {
		return ""
	}
	keys := make([]string, 0, len(t.EnvVars))
	for _, k := range t.EnvVars {
		if t.MCPCredentials[k] != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		b.WriteString(" ")
		b.WriteString(flag)
		b.WriteString(" ")
		b.WriteString(shellQuote(k + "=" + t.MCPCredentials[k]))
	}
	return b.String()
}

// shellQuote wraps a value in single quotes the POSIX way. Sufficient
// for MCP names, packages, URLs, and headers (none contain newlines in
// practice).
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func shellJoin(parts []string) string {
	out := make([]string, len(parts))
	for i, p := range parts {
		out[i] = shellQuote(p)
	}
	return strings.Join(out, " ")
}

// ---------------------------------------------------------------------
// claude-code adapter — `claude mcp add` (stdio | http | sse)
// Storage: ~/.claude.json (user scope).
// Docs: https://code.claude.com/docs/en/mcp
// ---------------------------------------------------------------------

type claudeCodeAdapter struct{}

func (claudeCodeAdapter) Name() string   { return "claude-code" }
func (claudeCodeAdapter) Binary() string { return "claude" }

func (claudeCodeAdapter) Register(ctx context.Context, t preset.Tool, out chan<- Line) error {
	name := mcpServerName(t)
	var cmd string
	switch mcpTransport(t) {
	case "stdio":
		cmd = fmt.Sprintf("claude mcp add --transport stdio --scope user%s %s -- %s",
			stdioEnvFlags(t, "--env"), shellQuote(name), shellJoin(stdioLaunchParts(t)))
	case "http":
		cmd = fmt.Sprintf("claude mcp add --transport http --scope user %s %s%s",
			shellQuote(name), shellQuote(expandCreds(t.MCPURL, t)), claudeHeaderFlags(t))
	case "sse":
		cmd = fmt.Sprintf("claude mcp add --transport sse --scope user %s %s%s",
			shellQuote(name), shellQuote(expandCreds(t.MCPURL, t)), claudeHeaderFlags(t))
	default:
		return fmt.Errorf("unsupported transport %q for claude-code", mcpTransport(t))
	}
	return runCmd(ctx, t.Name, cmd, out)
}

// claudeHeaderFlags renders --header "K: V" flags. Header values that
// expand to empty strings (because the credential wasn't collected) are
// dropped — sending `Authorization: Bearer ` would just trigger a 401
// that confuses the user more than a missing header.
func claudeHeaderFlags(t preset.Tool) string {
	if len(t.MCPHeaders) == 0 {
		return ""
	}
	keys := make([]string, 0, len(t.MCPHeaders))
	for k := range t.MCPHeaders {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		v := strings.TrimSpace(expandCreds(t.MCPHeaders[k], t))
		if v == "" {
			continue
		}
		b.WriteString(" --header ")
		b.WriteString(shellQuote(k + ": " + v))
	}
	return b.String()
}

// ---------------------------------------------------------------------
// codex adapter — `codex mcp add` (stdio | streamable-http; no SSE)
// Storage: ~/.codex/config.toml under [mcp.servers.<name>].
// Docs: https://developers.openai.com/codex/cli/reference
// ---------------------------------------------------------------------

type codexAdapter struct{}

func (codexAdapter) Name() string   { return "codex" }
func (codexAdapter) Binary() string { return "codex" }

func (codexAdapter) Register(ctx context.Context, t preset.Tool, out chan<- Line) error {
	name := mcpServerName(t)
	var cmd string
	switch mcpTransport(t) {
	case "stdio":
		cmd = fmt.Sprintf("codex mcp add%s %s -- %s",
			stdioEnvFlags(t, "--env"), shellQuote(name), shellJoin(stdioLaunchParts(t)))
	case "http":
		cmd = fmt.Sprintf("codex mcp add %s --url %s",
			shellQuote(name), shellQuote(expandCreds(t.MCPURL, t)))
	case "sse":
		// codex doesn't speak SSE — only stdio + streamable-http. Skip
		// quietly; the user will see this on the done screen if they
		// care.
		out <- Line{Tool: t.Name, Stream: "meta",
			Text: "skipping codex registration: codex does not support SSE transport"}
		return nil
	default:
		return fmt.Errorf("unsupported transport %q for codex", mcpTransport(t))
	}
	return runCmd(ctx, t.Name, cmd, out)
}

// ---------------------------------------------------------------------
// droid adapter — `droid mcp add` (stdio | http)
// Storage: ~/.factory/mcp.json with `mcpServers` map.
// Docs: https://docs.factory.ai/cli/configuration/mcp
// ---------------------------------------------------------------------

type droidAdapter struct{}

func (droidAdapter) Name() string   { return "droid" }
func (droidAdapter) Binary() string { return "droid" }

func (droidAdapter) Register(ctx context.Context, t preset.Tool, out chan<- Line) error {
	name := mcpServerName(t)
	var cmd string
	switch mcpTransport(t) {
	case "stdio":
		// droid's CLI takes the launch command as a single quoted
		// argument, NOT after `--`. e.g.
		//   droid mcp add hubspot "npx -y @hubspot/mcp-server"
		launch := strings.Join(stdioLaunchParts(t), " ")
		cmd = fmt.Sprintf("droid mcp add%s %s %s",
			stdioEnvFlags(t, "--env"), shellQuote(name), shellQuote(launch))
	case "http":
		cmd = fmt.Sprintf("droid mcp add %s %s --type http",
			shellQuote(name), shellQuote(expandCreds(t.MCPURL, t)))
	case "sse":
		// droid supports stdio + http only.
		out <- Line{Tool: t.Name, Stream: "meta",
			Text: "skipping droid registration: droid does not support SSE transport"}
		return nil
	default:
		return fmt.Errorf("unsupported transport %q for droid", mcpTransport(t))
	}
	return runCmd(ctx, t.Name, cmd, out)
}

// ---------------------------------------------------------------------
// opencode adapter — JSON file merge (no `opencode mcp add` CLI)
// Storage: ~/.config/opencode/opencode.json under top-level `mcp` map.
// Schema: { "mcp": { "<name>": { "type": "local"|"remote", ... } } }
// Docs: https://opencode.ai/docs/mcp-servers/
// ---------------------------------------------------------------------

type opencodeAdapter struct{}

func (opencodeAdapter) Name() string   { return "opencode" }
func (opencodeAdapter) Binary() string { return "opencode" }

func (opencodeAdapter) Register(ctx context.Context, t preset.Tool, out chan<- Line) error {
	cfgPath, err := opencodeConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		return fmt.Errorf("create opencode config dir: %w", err)
	}

	// Read existing config (or start fresh). We preserve every key on
	// the top level so a hand-edited config keeps its other fields.
	cfg := map[string]any{}
	if data, err := os.ReadFile(cfgPath); err == nil && len(data) > 0 {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("parse %s: %w", cfgPath, err)
		}
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", cfgPath, err)
	}

	mcpMap, _ := cfg["mcp"].(map[string]any)
	if mcpMap == nil {
		mcpMap = map[string]any{}
	}

	name := mcpServerName(t)
	entry := map[string]any{"enabled": true}
	switch mcpTransport(t) {
	case "stdio":
		entry["type"] = "local"
		// Convert []string → []any so json marshal preserves array type.
		parts := stdioLaunchParts(t)
		cmd := make([]any, len(parts))
		for i, p := range parts {
			cmd[i] = p
		}
		entry["command"] = cmd
		if env := credEnvMap(t); len(env) > 0 {
			entry["environment"] = env
		}
	case "http", "sse":
		entry["type"] = "remote"
		entry["url"] = expandCreds(t.MCPURL, t)
		if len(t.MCPHeaders) > 0 {
			h := map[string]any{}
			for k, v := range t.MCPHeaders {
				expanded := strings.TrimSpace(expandCreds(v, t))
				if expanded == "" {
					continue
				}
				h[k] = expanded
			}
			if len(h) > 0 {
				entry["headers"] = h
			}
		}
	default:
		return fmt.Errorf("unsupported transport %q for opencode", mcpTransport(t))
	}
	mcpMap[name] = entry
	cfg["mcp"] = mcpMap
	if _, ok := cfg["$schema"]; !ok {
		cfg["$schema"] = "https://opencode.ai/config.json"
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal opencode config: %w", err)
	}
	if err := os.WriteFile(cfgPath, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", cfgPath, err)
	}
	out <- Line{Tool: t.Name, Stream: "meta", Text: "wrote " + cfgPath}
	return nil
}

// credEnvMap returns the EnvVars-with-collected-values as a map suitable
// for embedding in a harness config (e.g. opencode's `environment`).
// Skips keys with empty values so the user's missing-credential reminder
// on the done screen still makes sense.
func credEnvMap(t preset.Tool) map[string]any {
	if len(t.EnvVars) == 0 || len(t.MCPCredentials) == 0 {
		return nil
	}
	out := map[string]any{}
	for _, k := range t.EnvVars {
		if v := t.MCPCredentials[k]; v != "" {
			out[k] = v
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func opencodeConfigPath() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "opencode", "opencode.json"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "opencode", "opencode.json"), nil
}
