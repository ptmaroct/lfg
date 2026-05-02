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
	"os"
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
	// "curl" is a friendly alias for the custom shell-script runner —
	// preset entries that install via `curl ... | sh` use Source="curl"
	// so the VIA column reads accurately instead of the catch-all "custom".
	registry["curl"] = registry["custom"]
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
//
// When a selected tool's name matches a bootstrap-able backend (e.g.
// the user picked the `mise` tool, and `mise` is also a backend that
// would otherwise self-bootstrap before its first dependent), we skip
// the auto-bootstrap step — installing mise twice in one run was a
// real bug seen in container testing where the explicit `mise` tool
// install ran via curl and the auto-bootstrap then ran the same curl
// again.
func Plan(bundles []preset.Bundle, selected map[string]bool) []Step {
	var steps []Step
	seenBootstrap := map[string]bool{}

	// Pre-scan: which bootstrap-capable backends are already covered by
	// an explicit selected tool of the same name?
	selfBootstrapped := map[string]bool{}
	for _, b := range bundles {
		for _, t := range b.Tools {
			if !selected[b.ID+"/"+t.Name] {
				continue
			}
			if needsBootstrap(t.Name) {
				selfBootstrapped[t.Name] = true
			}
		}
	}

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
			// Bootstrap once per backend that needs it AND isn't already
			// being installed as an explicit tool earlier in the queue.
			if needsBootstrap(backend) && !seenBootstrap[backend] && !selfBootstrapped[backend] {
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
	// Augment PATH so commands run in step N see binaries dropped by
	// step N-1 (e.g. mise → ~/.local/bin/mise). Restore on exit so we
	// don't leak PATH changes back to the caller.
	originalPath := os.Getenv("PATH")
	_ = os.Setenv("PATH", augmentedPath(originalPath))
	defer os.Setenv("PATH", originalPath)

	var failed []FailedStep
	failedBackends := map[string]bool{}
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

		// Re-augment PATH each iteration in case a prior step landed a
		// new binary into a directory we didn't have on PATH yet.
		_ = os.Setenv("PATH", augmentedPath(originalPath))

		// Skip-with-message when an upstream backend already failed
		// (e.g. mise bootstrap failed, all mise-source tools should
		// short-circuit instead of each one re-emitting the same
		// "mise: not found" exit-127 line).
		if !step.Bootstrap && failedBackends[step.Backend] {
			out <- Line{Tool: key, Stream: "begin"}
			msg := fmt.Sprintf("skipped — backend %q unavailable (upstream failed)", step.Backend)
			out <- Line{Tool: key, Stream: "meta", Text: msg}
			out <- Line{Tool: key, Stream: "end", Text: msg}
			failed = append(failed, FailedStep{Step: step, Err: fmt.Errorf("%s", msg)})
			continue
		}

		// Pre-check backend availability for non-bootstrap steps. Catches
		// the case where the backend's binary isn't on PATH (e.g. npm /
		// npx for skills before node is installed) without spending a
		// whole exec to learn that.
		if !step.Bootstrap && !i.Available() {
			out <- Line{Tool: key, Stream: "begin"}
			msg := fmt.Sprintf("skipped — %s not available on PATH (install its prerequisite first)", step.Backend)
			out <- Line{Tool: key, Stream: "meta", Text: msg}
			out <- Line{Tool: key, Stream: "end", Text: msg}
			failedBackends[step.Backend] = true
			failed = append(failed, FailedStep{Step: step, Err: fmt.Errorf("%s", msg)})
			continue
		}

		out <- Line{Tool: key, Stream: "begin"}
		var err error
		if step.Bootstrap {
			err = i.Bootstrap(ctx, out)
		} else {
			err = i.Install(ctx, step.Tool, out)
		}

		// Refresh PATH after the step in case it dropped a binary.
		_ = os.Setenv("PATH", augmentedPath(originalPath))

		// PostInstall: run each shell command sequentially after the
		// main install succeeds. Used for runtime prerequisites the
		// primary installer doesn't cover — e.g. agent-browser ships
		// as a skill stub but needs `npm i -g agent-browser` + `agent-
		// browser install` for the binary + Chrome bundle.
		//
		// Post-install steps are BEST EFFORT — a failure here (e.g.
		// Chrome for Testing has no linux/arm64 build) does NOT taint
		// the parent step's success. The main install (skill stub /
		// brew install / etc) already landed; the user can re-run or
		// patch manually. Surfacing post-install failures as full step
		// failures would falsely red-flag tools that are actually 80%
		// working, which is more confusing than a clear warning line.
		if err == nil && !step.Bootstrap && len(step.Tool.PostInstall) > 0 {
			for _, post := range step.Tool.PostInstall {
				if ctx.Err() != nil {
					break
				}
				out <- Line{Tool: key, Stream: "meta", Text: "post-install: " + post}
				if perr := runCmd(ctx, step.Tool.Name, post, out); perr != nil {
					out <- Line{Tool: key, Stream: "stderr",
						Text: fmt.Sprintf("post-install warning: %s — %v (main install still succeeded)", post, perr)}
				}
				_ = os.Setenv("PATH", augmentedPath(originalPath))
			}
		}

		errText := ""
		if err != nil {
			errText = err.Error()
			failed = append(failed, FailedStep{Step: step, Err: err})
			if step.Bootstrap {
				// A failed bootstrap means every dependent in this run
				// will also fail. Mark so we skip them with a clean msg.
				failedBackends[step.Backend] = true
			}
		}
		out <- Line{Tool: key, Stream: "end", Text: errText}
	}

	// Persist PATH augmentation into the user's shell rc so a fresh
	// shell sees the new bins without manual `source`. Idempotent — re-
	// runs replace the fenced block in place.
	if rc, err := EnsureShellPath(); err == nil && rc != "" {
		out <- Line{Tool: "shell", Stream: "meta",
			Text: fmt.Sprintf("updated %s — reload your shell or run `exec $SHELL`", rc)}
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
