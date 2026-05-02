package installer

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/ptmaroct/lfg/internal/preset"
)

// skillsInstaller handles Source="skills" — runs `npx skills add <url>
// --skill <name> -g -a <harness>... -y --copy` so a skill is installed
// globally into the harnesses detected on the host (Claude Code, Codex,
// OpenCode). Requires `npx` (i.e. Node), which is itself installed via
// the default bundle's mise step.
//
// Defaults locked:
//   - `-g` (global) — skills land in ~/.claude/skills/ (and equivalents)
//     so they're available to every project, matching lfg's "set up the
//     machine once" model.
//   - `--copy` — files are copied into the harness skills directory
//     rather than symlinked. Stable across upstream moves; doesn't break
//     when the source repo is deleted or moved.
//   - `-y` — skip every interactive prompt the skills CLI throws.
//
// The `-a` (agent) flags are populated dynamically via SetHarnesses,
// called by detect after probing the host. When no harness is detected
// we fall back to `claude-code` so the command stays valid.
type skillsInstaller struct{}

func (skillsInstaller) Name() string { return "skills" }

func (skillsInstaller) Available() bool {
	_, err := exec.LookPath("npx")
	return err == nil
}

func (s skillsInstaller) Bootstrap(ctx context.Context, out chan<- Line) error {
	if s.Available() {
		out <- Line{Tool: "skills", Stream: "meta", Text: "npx available"}
		return nil
	}
	out <- Line{Tool: "skills", Stream: "stderr",
		Text: "npx not found — install Node first (e.g. via mise: `mise use -g node@lts`)"}
	return errMissingNpx
}

func (s skillsInstaller) Install(ctx context.Context, t preset.Tool, out chan<- Line) error {
	cmdline := skillsCommand(t)
	if cmdline == "" {
		return fmt.Errorf("skill %q has no SkillURL", t.Name)
	}
	return runCmd(ctx, t.Name, cmdline, out)
}

func (s skillsInstaller) DryRun(t preset.Tool) string { return skillsCommand(t) }

// skillsCommand renders the npx invocation for one skill Tool. Returns
// empty string when SkillURL is missing — caller treats that as an error.
func skillsCommand(t preset.Tool) string {
	if t.SkillURL == "" {
		return ""
	}
	parts := []string{"npx", "skills", "add", t.SkillURL,
		"--skill", t.Name, "-g", "-y", "--copy"}
	for _, h := range getHarnesses() {
		parts = append(parts, "-a", h)
	}
	return strings.Join(parts, " ")
}

// SetHarnesses configures the agents the skills backend will pass via
// `-a` flags on each install. Empty input falls back to claude-code so
// `npx skills add` always has at least one target.
//
// Called by the CLI / TUI after detect.DetectedHarnesses() — both the
// headless apply path and the interactive probe screen plumb this so
// installs route to whatever harnesses are actually on the host.
func SetHarnesses(h []string) {
	skillHarnessesLock.Lock()
	defer skillHarnessesLock.Unlock()
	if len(h) == 0 {
		skillHarnesses = []string{"claude-code"}
		return
	}
	skillHarnesses = append(skillHarnesses[:0:0], h...)
}

func getHarnesses() []string {
	skillHarnessesLock.RLock()
	defer skillHarnessesLock.RUnlock()
	return append([]string(nil), skillHarnesses...)
}

var (
	skillHarnesses     = []string{"claude-code"}
	skillHarnessesLock sync.RWMutex
)

var errMissingNpx = errors.New("npx not installed")
