package installer

import (
	"context"
	"os/exec"

	"github.com/anuj/lfg/internal/preset"
)

// brewInstaller handles Source="brew" and Source="cask" tools.
//
// Bootstrap installs Homebrew via the official one-liner. Mac is the
// primary path; Linux gets linuxbrew via the same script (works in our
// Ubuntu Docker image but we expect most Linux users to use apt).
type brewInstaller struct{}

func (brewInstaller) Name() string { return "brew" }

func (brewInstaller) Available() bool {
	_, err := exec.LookPath("brew")
	return err == nil
}

const brewBootstrapCmd = `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`

func (b brewInstaller) Bootstrap(ctx context.Context, out chan<- Line) error {
	if b.Available() {
		out <- Line{Tool: "brew", Stream: "meta", Text: "brew already installed"}
		return nil
	}
	return runCmd(ctx, "brew", brewBootstrapCmd, out)
}

func (b brewInstaller) Install(ctx context.Context, t preset.Tool, out chan<- Line) error {
	cmd := installCmd(t)
	return runCmd(ctx, t.Name, cmd, out)
}

func (b brewInstaller) DryRun(t preset.Tool) string { return installCmd(t) }
