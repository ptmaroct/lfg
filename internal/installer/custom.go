package installer

import (
	"context"
	"errors"

	"github.com/ptmaroct/lfg/internal/preset"
)

// customInstaller handles Source="custom" — raw shell command from the
// preset, e.g. `curl -sS https://starship.rs/install.sh | sh -s -- -y`.
// No bootstrap required: by definition the command is self-contained.
type customInstaller struct{}

func (customInstaller) Name() string { return "custom" }

// Always available — `sh` ships with every supported OS.
func (customInstaller) Available() bool { return true }

func (customInstaller) Bootstrap(ctx context.Context, out chan<- Line) error { return nil }

func (customInstaller) Install(ctx context.Context, t preset.Tool, out chan<- Line) error {
	cmd := installCmd(t)
	if t.PinSHA256 != "" {
		return runVerifiedCurl(ctx, t.Name, cmd, t.PinSHA256, out)
	}
	if t.Source == "curl" {
		out <- Line{Tool: t.Name, Stream: "meta",
			Text: "WARN: no PinSHA256 — installing without supply-chain verification"}
	}
	return runCmd(ctx, t.Name, cmd, out)
}

func (customInstaller) DryRun(t preset.Tool) string { return installCmd(t) }

// errMissingNpm is shared with npm.go.
var errMissingNpm = errors.New("npm not installed")
