package installer

import (
	"context"
	"os/exec"
	"runtime"

	"github.com/ptmaroct/lfg/internal/preset"
)

// miseInstaller handles Source="mise" — language version manager.
// Bootstrap pulls mise via brew on Mac (if available) or the official
// install script otherwise.
type miseInstaller struct{}

func (miseInstaller) Name() string { return "mise" }

func (miseInstaller) Available() bool {
	_, err := exec.LookPath("mise")
	return err == nil
}

const miseInstallScript = `curl -fsSL https://mise.run | sh`

func (m miseInstaller) Bootstrap(ctx context.Context, out chan<- Line) error {
	if m.Available() {
		out <- Line{Tool: "mise", Stream: "meta", Text: "mise already installed"}
		return nil
	}
	cmd := miseInstallScript
	if runtime.GOOS == "darwin" {
		// Prefer brew on Mac when available — keeps mise upgrade story
		// consistent with the rest of the brew-managed tools.
		if _, err := exec.LookPath("brew"); err == nil {
			cmd = "brew install mise"
		}
	}
	return runCmd(ctx, "mise", cmd, out)
}

func (m miseInstaller) Install(ctx context.Context, t preset.Tool, out chan<- Line) error {
	return runCmd(ctx, t.Name, installCmd(t), out)
}

func (m miseInstaller) DryRun(t preset.Tool) string { return installCmd(t) }
