package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// shellRCMarker fences the lfg-managed block in the user's shell rc.
// Idempotency hinges on this exact string — never change it without a
// migration path that prunes old blocks first.
const shellRCMarker = "# lfg-managed PATH (do not edit between markers)"
const shellRCEndMarker = "# end lfg-managed PATH"

// EnsureShellPath appends an idempotent PATH export to the user's shell
// rc file so freshly installed tools (mise, brew, claude, etc.) survive
// the next shell. Returns the rc path written, or empty if no change /
// no rc detected. Errors are returned but callers can usually ignore
// them — failing to write rc is annoying but not fatal.
//
// The block is bounded by markers so re-running lfg replaces it
// in-place rather than appending duplicates each time.
func EnsureShellPath() (string, error) {
	rc := detectShellRC()
	if rc == "" {
		return "", nil
	}
	block := buildShellBlock()
	existing, _ := os.ReadFile(rc)
	updated, changed := upsertManagedBlock(string(existing), block)
	if !changed {
		return rc, nil
	}
	if err := os.WriteFile(rc, []byte(updated), 0o644); err != nil {
		return rc, fmt.Errorf("write %s: %w", rc, err)
	}
	return rc, nil
}

// detectShellRC picks the rc file matching $SHELL. macOS bash needs
// ~/.bash_profile (login shell); Linux bash uses ~/.bashrc. Falls back
// to .zshrc when shell can't be parsed since zsh is the macOS default
// and lfg's primary target.
func detectShellRC() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	shell := strings.ToLower(filepath.Base(os.Getenv("SHELL")))
	switch {
	case strings.Contains(shell, "zsh"):
		return filepath.Join(home, ".zshrc")
	case strings.Contains(shell, "bash"):
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, ".bash_profile")
		}
		return filepath.Join(home, ".bashrc")
	case strings.Contains(shell, "fish"):
		return filepath.Join(home, ".config", "fish", "config.fish")
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
