package installer

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/ptmaroct/lfg/internal/preset"
)

// shellAliasMarker fences the lfg-managed alias block in the user's
// shell rc. Distinct from shellRCMarker so PATH and aliases live in
// independent fenced regions and can be replaced/removed in isolation.
const shellAliasMarker = "# lfg-managed aliases (do not edit between markers)"
const shellAliasEndMarker = "# end lfg-managed aliases"

// EnsureShellAliases writes the user's selected aliases as a fenced
// block to every detected shell rc (bash, zsh, fish). Idempotent: same
// inputs produce the same block. Empty input → block is removed
// (deselecting every alias should leave no orphan markers behind).
//
// Returns the comma-joined list of paths actually touched and the
// first per-file write error (if any) for surface-to-CLI logging.
func EnsureShellAliases(aliases []preset.Alias) (string, error) {
	targets := candidateRCs()
	if len(targets) == 0 {
		return "", nil
	}
	var written []string
	var firstErr error
	for _, rc := range targets {
		existing, _ := os.ReadFile(rc)
		block := buildAliasBlock(rc, aliases)
		updated, changed := upsertManagedBlock(string(existing), block, shellAliasMarker, shellAliasEndMarker)
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

// buildAliasBlock renders the rc-specific alias snippet. Returns an
// empty string when there are no aliases to write so upsertManagedBlock
// strips the prior block in place.
func buildAliasBlock(rcPath string, aliases []preset.Alias) string {
	if len(aliases) == 0 {
		return ""
	}
	isFish := strings.HasSuffix(rcPath, "config.fish")
	var lines []string
	lines = append(lines, shellAliasMarker)
	for _, a := range aliases {
		line := renderAliasLine(rcPath, a, isFish)
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	lines = append(lines, shellAliasEndMarker)
	if len(lines) == 2 {
		// All aliases skipped → no block.
		return ""
	}
	return strings.Join(lines, "\n")
}

// renderAliasLine emits one rc line for the given alias.
//
// `reload` is special-cased per shell: bash/zsh resolve to a self-exec
// of the matching shell so PATH/alias edits take effect; fish needs a
// function (a fish alias of `exec fish` doesn't survive the exec).
func renderAliasLine(rcPath string, a preset.Alias, isFish bool) string {
	cmd := a.Command
	if isFish && a.FishCommand != "" {
		cmd = a.FishCommand
	}
	if a.Name == "reload" {
		switch {
		case isFish:
			return "function reload; exec fish; end"
		case strings.HasSuffix(rcPath, ".zshrc"):
			return "alias reload='exec zsh'"
		default:
			return "alias reload='exec bash'"
		}
	}
	if cmd == "" || a.Name == "" {
		return ""
	}
	if isFish {
		return fmt.Sprintf("alias %s %s", a.Name, fishQuote(cmd))
	}
	return fmt.Sprintf("alias %s=%s", a.Name, posixSingleQuote(cmd))
}

// posixSingleQuote wraps s in single quotes for bash/zsh. Embedded
// single quotes are escaped as '\'' (close-quote, escaped quote,
// reopen-quote) — the canonical POSIX trick.
func posixSingleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// fishQuote wraps s in single quotes for fish. Fish's quoting is
// simpler than POSIX: backslash-escape internal singles.
func fishQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `\'`) + "'"
}

// existingAliasRE matches `alias NAME=...` (bash/zsh) and
// `alias NAME ...` (fish). Captures the alias name. Optional leading
// whitespace; ignores commented lines (handled by caller before regex
// match).
var existingAliasRE = regexp.MustCompile(`^\s*alias\s+([A-Za-z_][A-Za-z0-9_-]*)\s*[= ]`)

// ExistingAliases scans every candidate rc file for `alias NAME=...`
// definitions OUTSIDE the lfg-managed fenced block. Returned map is
// keyed by alias name with value `<rcPath>:<lineNumber>` so the picker
// can render conflict warnings (`gd ⚠ defined in .zshrc:42`). Lines
// inside our own fenced block are intentionally skipped: we own those.
//
// Best-effort. Read errors are silently swallowed so a missing rc
// doesn't break the picker — empty map just means "no conflicts found".
func ExistingAliases() map[string]string {
	out := map[string]string{}
	for _, rc := range candidateRCs() {
		f, err := os.Open(rc)
		if err != nil {
			continue
		}
		short := shortRCName(rc)
		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		lineNo := 0
		insideManaged := false
		for scanner.Scan() {
			lineNo++
			line := scanner.Text()
			trimmed := strings.TrimSpace(line)
			if strings.Contains(trimmed, shellAliasMarker) {
				insideManaged = true
				continue
			}
			if strings.Contains(trimmed, shellAliasEndMarker) {
				insideManaged = false
				continue
			}
			if insideManaged {
				continue
			}
			if strings.HasPrefix(trimmed, "#") {
				continue
			}
			m := existingAliasRE.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			name := m[1]
			if _, seen := out[name]; seen {
				continue
			}
			out[name] = fmt.Sprintf("%s:%d", short, lineNo)
		}
		_ = f.Close()
	}
	return out
}

// shortRCName returns just the basename of an rc path so conflict
// messages stay terse (`/home/u/.bashrc` → `.bashrc`).
func shortRCName(p string) string {
	return filepath.Base(p)
}

// SortedAliasNames is a tiny helper used by the TUI to render aliases
// in a stable order across snapshots. Caller-controlled to keep the
// picker model free of unrelated logic.
func SortedAliasNames(aliases []preset.Alias) []string {
	out := make([]string, 0, len(aliases))
	for _, a := range aliases {
		out = append(out, a.Name)
	}
	sort.Strings(out)
	return out
}
