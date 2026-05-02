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
//
// Harness list is resolved LIVE at command-build time (not from the
// cached probe at TUI startup). Reason: when the user installs codex /
// opencode in the same run as their skills, detect at TUI startup ran
// before those CLIs existed and cached an incomplete harness list —
// the resulting `npx skills add` ran with only `-a claude-code` and
// the skills never landed in `~/.codex/skills`. Re-probing here picks
// up the harnesses already installed by earlier steps in this run.
func skillsCommand(t preset.Tool) string {
	if t.SkillURL == "" {
		return ""
	}
	parts := []string{"npx", "skills", "add", t.SkillURL,
		"--skill", t.Name, "-g", "-y", "--copy"}
	harnesses := liveHarnesses()
	if len(harnesses) == 0 {
		// Nothing on PATH right now — fall back to cached list (may
		// have been seeded by detect or by an explicit SetHarnesses
		// from tests).
		harnesses = getHarnesses()
	}
	for _, h := range harnesses {
		parts = append(parts, "-a", h)
	}
	return strings.Join(parts, " ")
}

// liveHarnesses probes for AI-harness CLIs on PATH right now. Mirrors
// the table in internal/detect/detect.go but lives here too to avoid an
// installer→detect import (detect already imports preset; we keep the
// dep graph one-way). Keep the lists in sync.
func liveHarnesses() []string {
	candidates := []struct{ agent, bin string }{
		{"claude-code", "claude"},
		{"codex", "codex"},
		{"opencode", "opencode"},
	}
	var out []string
	for _, c := range candidates {
		if _, err := exec.LookPath(c.bin); err == nil {
			out = append(out, c.agent)
		}
	}
	return out
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
