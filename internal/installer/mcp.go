package installer

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/ptmaroct/lfg/internal/preset"
)

// mcpInstaller handles Source="mcp" — Model Context Protocol servers.
//
// Two transports are supported on the schema side:
//   - stdio (default): a local subprocess spawned by the harness, usually
//     `npx -y <package>`. We pre-install the npm package globally so the
//     first launch doesn't pay the npx download cost.
//   - http / sse: remote endpoints. Nothing to install locally; we just
//     register the URL with each target harness.
//
// After install, we walk the tool's TargetHarnesses list and ask each
// matching adapter (registered in mcp_register.go) to wire the server
// up. Adapters skip themselves automatically when their CLI binary is
// not on PATH — installing a harness later and re-running `lfg apply`
// picks them up.
//
// Adding a new harness to lfg = drop a new adapter into mcp_register.go
// and add its name to KnownHarnesses(); existing MCP entries pick it
// up via the `target_harnesses = ["all"]` shorthand or by listing the
// new harness explicitly.
type mcpInstaller struct{}

func (mcpInstaller) Name() string { return "mcp" }

func (mcpInstaller) Available() bool {
	// stdio MCPs need npm to install the package; remote MCPs don't.
	// Worst case for a remote-only entry: we'd skip it because npm is
	// missing, even though no install was actually needed. The bundled
	// defaults are all stdio so this is a non-issue for v0.1; refine
	// when remote-by-default MCPs ship.
	_, err := exec.LookPath("npm")
	return err == nil
}

func (m mcpInstaller) Bootstrap(ctx context.Context, out chan<- Line) error {
	if m.Available() {
		out <- Line{Tool: "mcp", Stream: "meta", Text: "npm available"}
		return nil
	}
	out <- Line{Tool: "mcp", Stream: "stderr",
		Text: "npm not found — install Node first (e.g. via mise: `mise use -g node@lts`)"}
	return errMissingNpmForMCP
}

func (m mcpInstaller) Install(ctx context.Context, t preset.Tool, out chan<- Line) error {
	switch mcpTransport(t) {
	case "stdio":
		if t.MCPPackage == "" {
			return fmt.Errorf("mcp tool %q has no mcp_package", t.Name)
		}
		if err := runCmd(ctx, t.Name, "npm install -g "+t.MCPPackage, out); err != nil {
			return err
		}
	case "http", "sse":
		if t.MCPURL == "" {
			return fmt.Errorf("mcp tool %q has no mcp_url for transport %q", t.Name, mcpTransport(t))
		}
		out <- Line{Tool: t.Name, Stream: "meta",
			Text: fmt.Sprintf("remote MCP — nothing to install locally (url: %s)", t.MCPURL)}
	default:
		return fmt.Errorf("mcp tool %q has unknown mcp_type %q", t.Name, t.MCPType)
	}

	// Wire into every requested harness adapter that's actually present
	// on the host. Failures here surface as warnings — the package /
	// remote URL is already valid; registration is fixable manually.
	registerWithHarnesses(ctx, t, out)
	return nil
}

func (m mcpInstaller) DryRun(t preset.Tool) string {
	switch mcpTransport(t) {
	case "stdio":
		if t.MCPPackage == "" {
			return ""
		}
		return "npm install -g " + t.MCPPackage
	case "http", "sse":
		return "# remote MCP " + t.MCPURL + " — register with target harnesses"
	}
	return ""
}

// mcpTransport returns the effective transport for the tool. Empty
// MCPType is treated as "stdio" for backward compat.
func mcpTransport(t preset.Tool) string {
	if t.MCPType == "" {
		return "stdio"
	}
	return t.MCPType
}

// registerWithHarnesses walks the tool's TargetHarnesses list and
// dispatches to each registered adapter whose CLI is on PATH.
//
// Special value "all" (or an empty list, for ergonomics) means
// "register with every harness lfg knows about" — adding a new harness
// to the registry then auto-extends every existing MCP entry.
//
// A failed registration emits a warning + the literal command the user
// can run manually; it does NOT fail the parent install step (the npm
// package is already on disk, so retrying the registration alone is
// safe).
func registerWithHarnesses(ctx context.Context, t preset.Tool, out chan<- Line) {
	targets := t.TargetHarnesses
	if len(targets) == 0 || containsString(targets, "all") {
		targets = KnownHarnesses()
	}
	for _, name := range targets {
		a, ok := harnessAdapters[name]
		if !ok {
			out <- Line{Tool: t.Name, Stream: "stderr",
				Text: fmt.Sprintf("unknown harness %q in target_harnesses (known: %v)", name, KnownHarnesses())}
			continue
		}
		if _, err := exec.LookPath(a.Binary()); err != nil {
			out <- Line{Tool: t.Name, Stream: "meta",
				Text: fmt.Sprintf("skipping %s registration: %s not on PATH (install %s first)",
					a.Name(), a.Binary(), a.Name())}
			continue
		}
		out <- Line{Tool: t.Name, Stream: "meta", Text: "registering with " + a.Name()}
		if err := a.Register(ctx, t, out); err != nil {
			out <- Line{Tool: t.Name, Stream: "stderr",
				Text: fmt.Sprintf("warning: %s registration failed — %v (re-run with `lfg apply` after fixing)", a.Name(), err)}
		}
	}
}

func containsString(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

var errMissingNpmForMCP = errors.New("npm not installed (required for mcp)")
