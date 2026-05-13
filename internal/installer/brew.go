package installer

import (
	"context"
	"os/exec"

	"github.com/ptmaroct/lfg/internal/preset"
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
	// Use the pinned SHA256 from pins.toml when available so the
	// bootstrap path matches the explicit-brew-tool install path. If
	// the pin is missing (older binary, hand-rolled config) fall back
	// to the un-verified curl-bash command for usability.
	if sha := preset.CurrentPins().Pins["barebones/brew"].SHA256; sha != "" {
		return runVerifiedCurl(ctx, "brew", brewBootstrapCmd, sha, out)
	}
	out <- Line{Tool: "brew", Stream: "meta",
		Text: "WARN: no PinSHA256 for brew installer — running unverified"}
	return runCmd(ctx, "brew", brewBootstrapCmd, out)
}

func (b brewInstaller) Install(ctx context.Context, t preset.Tool, out chan<- Line) error {
	cmd := installCmd(t)
	return runCmd(ctx, t.Name, cmd, out)
}

func (b brewInstaller) DryRun(t preset.Tool) string { return installCmd(t) }
