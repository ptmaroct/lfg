package installer

import (
	"context"
	"os/exec"

	"github.com/anuj/lfg/internal/preset"
)

// npmInstaller handles Source="npm" — global npm packages (codex,
// claude-code, opencode). Requires Node which lands via mise (so the
// AI-CLIs bundle is implicitly gated on the default bundle, but we
// don't enforce a hard ordering — the install will fail with a clear
// message if `npm` is missing).
type npmInstaller struct{}

func (npmInstaller) Name() string { return "npm" }

func (npmInstaller) Available() bool {
	_, err := exec.LookPath("npm")
	return err == nil
}

func (n npmInstaller) Bootstrap(ctx context.Context, out chan<- Line) error {
	if n.Available() {
		out <- Line{Tool: "npm", Stream: "meta", Text: "npm available"}
		return nil
	}
	out <- Line{Tool: "npm", Stream: "stderr",
		Text: "npm not found — install Node first (e.g. via mise: `mise use -g node@lts`)"}
	return errMissingNpm
}

func (n npmInstaller) Install(ctx context.Context, t preset.Tool, out chan<- Line) error {
	return runCmd(ctx, t.Name, installCmd(t), out)
}

func (n npmInstaller) DryRun(t preset.Tool) string { return installCmd(t) }
