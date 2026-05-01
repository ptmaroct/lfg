package installer

import (
	"context"
	"os/exec"

	"github.com/ptmaroct/lfg/internal/preset"
)

// aptInstaller handles Source="apt". No bootstrap needed — apt is on
// every Debian/Ubuntu by default. Tools assume sudo is available.
type aptInstaller struct{}

func (aptInstaller) Name() string { return "apt" }

func (aptInstaller) Available() bool {
	_, err := exec.LookPath("apt-get")
	return err == nil
}

func (aptInstaller) Bootstrap(ctx context.Context, out chan<- Line) error {
	out <- Line{Tool: "apt", Stream: "meta", Text: "apt is part of the base system; nothing to bootstrap"}
	return nil
}

func (aptInstaller) Install(ctx context.Context, t preset.Tool, out chan<- Line) error {
	return runCmd(ctx, t.Name, installCmd(t), out)
}

func (aptInstaller) DryRun(t preset.Tool) string { return installCmd(t) }
