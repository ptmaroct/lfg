package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// shellRCMarker fences the lfg-managed block in the user's shell rc.
// Idempotency hinges on this exact string — never change it without a
// migration path that prunes old blocks first.
const shellRCMarker = "# lfg-managed PATH (do not edit between markers)"
const shellRCEndMarker = "# end lfg-managed PATH"

// EnsureShellPath appends an idempotent PATH export to EVERY shell rc
// the user is likely to use — bash, zsh, fish — not just the active
// one. Reason: a developer routinely switches shells (e.g. lfg run
// from bash, tomorrow they `chsh -s zsh`), and we want freshly
// installed tools (mise, brew, claude, etc.) to survive that switch
// without re-running lfg. Idempotent — fenced block replaces in place
// per file. Returns the list of paths actually touched. Errors per
// file are joined into a single multi-error so partial-write isn't
// silently swallowed; callers can still treat the func as best-effort.
func EnsureShellPath() (string, error) {
	targets := candidateRCs()
	if len(targets) == 0 {
		return "", nil
	}
	block := buildShellBlock()
	var written []string
	var firstErr error
	for _, rc := range targets {
		existing, _ := os.ReadFile(rc)
		updated, changed := upsertManagedBlock(string(existing), block)
		if !changed {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(rc), 0o755); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("mkdir %s: %w", filepath.Dir(rc), err)
			continue
		}
		if err := os.WriteFile(rc, []byte(updated), 0o644); err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("write %s: %w", rc, err)
			}
			continue
		}
		written = append(written, rc)
	}
	return strings.Join(written, ", "), firstErr
}

// candidateRCs returns every shell rc lfg should keep PATH-augmented.
// We always touch the current shell's rc (so the immediate
// post-install reload works) plus any other common rcs that already
// exist on disk (so a future `chsh` doesn't lose the augmentation).
// Fish config dir is auto-created on write since fish stores it under
// ~/.config/fish/ which may not exist yet on a fresh box.
func candidateRCs() []string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return nil
	}
	wanted := map[string]bool{
		detectShellRC(): true,
	}
	// Also include common rcs that already exist so an existing zsh
	// user picking up lfg from a bash session keeps both rcs in sync.
	for _, name := range []string{".bashrc", ".zshrc"} {
		p := filepath.Join(home, name)
		if _, err := os.Stat(p); err == nil {
			wanted[p] = true
		}
	}
	var out []string
	for p := range wanted {
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	// Stable order so the comma-joined return string is deterministic.
	sort.Strings(out)
	return out
}

// detectShellRC picks the rc file matching the user's actual shell.
//
// Detection priority:
//  1. $SHELL env var — set by every well-behaved shell
//  2. /proc/<ppid>/comm parent process name — covers `docker run -it
//     bash` where $SHELL stays unset and we'd otherwise wrongly write
//     to ~/.zshrc (real bug observed in container testing)
//  3. fall back to whichever rc file exists on disk
//  4. last resort: ~/.bashrc on linux, ~/.zshrc on darwin
//
// We always target the INTERACTIVE rc (~/.bashrc, not ~/.bash_profile)
// so the matching `exec bash` reload command actually picks up the new
// PATH — login-shell bash on Linux sources .bash_profile, not .bashrc.
func detectShellRC() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	shell := strings.ToLower(filepath.Base(os.Getenv("SHELL")))
	if shell == "" || shell == "." {
		if data, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", os.Getppid())); err == nil {
			shell = strings.ToLower(strings.TrimSpace(string(data)))
		}
	}
	switch {
	case strings.Contains(shell, "zsh"):
		return filepath.Join(home, ".zshrc")
	case strings.Contains(shell, "bash"):
		return filepath.Join(home, ".bashrc")
	case strings.Contains(shell, "fish"):
		return filepath.Join(home, ".config", "fish", "config.fish")
	}
	for _, name := range []string{".bashrc", ".zshrc", ".profile"} {
		p := filepath.Join(home, name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	if runtime.GOOS == "linux" {
		return filepath.Join(home, ".bashrc")
	}
	return filepath.Join(home, ".zshrc")
}

// buildShellBlock returns the rc snippet. We export each known bin dir
// only when the dir actually exists at write time — avoids polluting
// PATH with phantom directories that confuse other tools.
func buildShellBlock() string {
	home, _ := os.UserHomeDir()
	var lines []string
	lines = append(lines, shellRCMarker)
	for _, d := range commonToolBinDirs(home) {
		if d == "" {
			continue
		}
		if _, err := os.Stat(d); err != nil {
			continue
		}
		lines = append(lines, fmt.Sprintf(`export PATH="%s:$PATH"`, d))
	}
	// brew shellenv generates the full set of HOMEBREW_* vars + PATH.
	// Only emit when brew exists; safe to source the missing path
	// otherwise becomes a startup error every shell.
	if home != "" {
		brewBin := "/home/linuxbrew/.linuxbrew/bin/brew"
		if runtime.GOOS == "darwin" {
			brewBin = "/opt/homebrew/bin/brew"
		}
		if _, err := os.Stat(brewBin); err == nil {
			lines = append(lines, fmt.Sprintf(`eval "$(%s shellenv)"`, brewBin))
		}
	}
	lines = append(lines, shellRCEndMarker)
	return strings.Join(lines, "\n")
}

// upsertManagedBlock replaces an existing fenced block in `existing`
// with `block`, or appends one if absent. Returns the new file content
// and a bool indicating whether anything changed.
func upsertManagedBlock(existing, block string) (string, bool) {
	startIdx := strings.Index(existing, shellRCMarker)
	endIdx := strings.Index(existing, shellRCEndMarker)
	if startIdx == -1 || endIdx == -1 || endIdx < startIdx {
		// Append. Ensure single newline between old content + block.
		out := strings.TrimRight(existing, "\n")
		if out != "" {
			out += "\n\n"
		}
		out += block + "\n"
		return out, true
	}
	// Replace in-place.
	endIdx += len(shellRCEndMarker)
	rebuilt := existing[:startIdx] + block + existing[endIdx:]
	if rebuilt == existing {
		return existing, false
	}
	return rebuilt, true
}
