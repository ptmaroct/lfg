// Package installer turns preset.Tool entries into actual install
// commands run on the host. Five backends are wired: brew, apt, mise,
// npm, and a generic "custom" shell-script runner. Output streams as
// Line values so the TUI can tail it live.
//
// Concurrency model: a single goroutine walks the resolved Plan,
// invoking the matching backend's Install for each step. Multiple
// installs in parallel are intentionally avoided — package managers
// don't reliably support concurrent invocations and serial output
// reads cleaner in the log tail.
package installer

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/ptmaroct/lfg/internal/preset"
)

// Line is one chunk of installer output bound for the TUI / logs.
type Line struct {
	Tool   string // bundle/tool key, e.g. "default/git"
	Stream string // "stdout" | "stderr" | "meta"
	Text   string
}

// Step is one entry in a resolved install plan. Bootstrap=true means
// "install the installer itself" (e.g. install brew before any brew
// formula). The tool field is empty for bootstrap steps.
type Step struct {
	Bundle    string
	Tool      preset.Tool
	Backend   string // "brew" / "apt" / "mise" / ...
	Bootstrap bool
}

// Installer is implemented per backend.
type Installer interface {
	Name() string
	// Available returns true when the installer binary is present.
	Available() bool
	// Bootstrap installs the installer itself if missing. Idempotent —
	// safe to call when Available() is already true.
	Bootstrap(ctx context.Context, out chan<- Line) error
	// Install runs the tool's install command for the host OS. Returns
	// an error on non-zero exit; output streams to `out`.
	Install(ctx context.Context, tool preset.Tool, out chan<- Line) error
	// DryRun returns the shell command line that would be run, without
	// running it. Empty string when no command is defined for this OS.
	DryRun(tool preset.Tool) string
}

// registry maps Source values to installer impls. Built lazily so
// individual backends can be swapped in tests.
var registry = map[string]Installer{}

// Register inserts (or replaces) a backend in the dispatch table.
// Tests use this to swap real implementations for fakes.
func Register(i Installer) { registry[i.Name()] = i }

// For returns the backend handling the given Source, or the custom
// runner as a catch-all when nothing matches.
func For(source string) Installer {
	if i, ok := registry[source]; ok {
		return i
	}
	return registry["custom"]
}

func init() {
	// Default registrations — tests may overwrite via Register.
	Register(&brewInstaller{})
	Register(&aptInstaller{})
	Register(&miseInstaller{})
	Register(&npmInstaller{})
	Register(&customInstaller{})
	Register(&skillsInstaller{})
	// brew handles both formulae and casks; "cask" routes through brew.
	registry["cask"] = registry["brew"]
}

// installCmd picks the right install command for the host OS. Returns
// an empty string when the tool has no command defined for this platform
// — caller should treat that as a no-op + meta line.
func installCmd(t preset.Tool) string {
	if runtime.GOOS == "darwin" {
		return strings.TrimSpace(t.InstallMac)
	}
	// Treat everything non-darwin as Linux for now. Windows is out of
	// scope (see plan). BSDs would route here too — reasonable default.
	return strings.TrimSpace(t.InstallLinux)
}

// Plan turns selected tools into an ordered list of Steps with the
// minimum bootstraps prepended. A backend appears as a bootstrap step
// at most once.
func Plan(bundles []preset.Bundle, selected map[string]bool) []Step {
	var steps []Step
	seenBootstrap := map[string]bool{}

	for _, b := range bundles {
		for _, t := range b.Tools {
			key := b.ID + "/" + t.Name
			if !selected[key] {
				continue
			}
			backend := t.Source
			if _, ok := registry[backend]; !ok {
				backend = "custom"
			}
			// Bootstrap once per backend that needs it.
			if needsBootstrap(backend) && !seenBootstrap[backend] {
				steps = append(steps, Step{Backend: backend, Bootstrap: true})
				seenBootstrap[backend] = true
			}
			steps = append(steps, Step{
				Bundle:  b.ID,
				Tool:    t,
				Backend: backend,
			})
		}
	}
	return steps
}

// needsBootstrap returns true when the backend's binary is something we
// can install ourselves (brew, mise) vs. a stdlib-or-package-manager
// dep we expect to be present (apt is on every Debian/Ubuntu).
func needsBootstrap(backend string) bool {
	switch backend {
	case "brew", "mise":
		return true
	}
	return false
}

// Run executes a Plan serially. For each step it picks the right
// backend and either bootstraps it or installs the tool. Returns the
// list of failed steps (zero-length on success). Cancellation via ctx.
//
// Emits two structural markers on `out` per step so consumers (TUI
// progress bar) can advance without parsing meta strings:
//   - {Stream:"begin", Tool: key}
//   - {Stream:"end",   Tool: key, Text: <error string or "">}
//
// Where `key` is "<backend>" for bootstraps and "<bundle>/<tool>"
// otherwise.
func Run(ctx context.Context, plan []Step, out chan<- Line) []FailedStep {
	var failed []FailedStep
	for _, step := range plan {
		if ctx.Err() != nil {
			break
		}
		key := stepKey(step)
		i := For(step.Backend)
		if i == nil {
			out <- Line{Tool: key, Stream: "begin"}
			out <- Line{Tool: key, Stream: "end", Text: "no installer"}
			failed = append(failed, FailedStep{Step: step, Err: fmt.Errorf("no installer for %q", step.Backend)})
			continue
		}
		out <- Line{Tool: key, Stream: "begin"}
		var err error
		if step.Bootstrap {
			err = i.Bootstrap(ctx, out)
		} else {
			err = i.Install(ctx, step.Tool, out)
		}
		errText := ""
		if err != nil {
			errText = err.Error()
			failed = append(failed, FailedStep{Step: step, Err: err})
		}
		out <- Line{Tool: key, Stream: "end", Text: errText}
	}
	return failed
}

// stepKey is the canonical identifier used in begin/end markers.
func stepKey(s Step) string {
	if s.Bootstrap {
		return s.Backend
	}
	return s.Bundle + "/" + s.Tool.Name
}

// FailedStep pairs a step with its terminal error for post-run reporting.
type FailedStep struct {
	Step Step
	Err  error
}
