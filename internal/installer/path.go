package installer

import (
	"os"
	"path/filepath"
	"strings"
)

// commonToolBinDirs lists locations that tools we install commonly land
// binaries into. Adding these to PATH at the start of a run (and after
// each step) is what lets a `mise use -g node@lts` step in step N find
// `mise` on PATH after step N-1 dropped it into ~/.local/bin — the
// child sh subprocess inherits whatever the parent process has, so
// without this it'd never see freshly installed tools.
//
// Order matters: leftmost wins on duplicate binary names. We prefer
// user-local installs over system fallbacks.
func commonToolBinDirs(home string) []string {
	if home == "" {
		return nil
	}
	return []string{
		filepath.Join(home, ".local", "bin"),
		filepath.Join(home, ".local", "share", "mise", "shims"),
		filepath.Join(home, ".cargo", "bin"),
		filepath.Join(home, ".bun", "bin"),
		filepath.Join(home, ".linuxbrew", "bin"),
		"/home/linuxbrew/.linuxbrew/bin",
		"/opt/homebrew/bin",
		"/usr/local/bin",
	}
}

// augmentedPath returns the existing PATH with our well-known dev bin
// dirs prepended (de-duplicated). When a dir doesn't exist on disk yet
// we still include it so a path lookup picks it up after the matching
// installer step lands a binary into it.
func augmentedPath(currentPath string) string {
	home, _ := os.UserHomeDir()
	add := commonToolBinDirs(home)

	seen := map[string]bool{}
	var parts []string
	for _, d := range add {
		if d == "" || seen[d] {
			continue
		}
		seen[d] = true
		parts = append(parts, d)
	}
	for _, d := range filepath.SplitList(currentPath) {
		if d == "" || seen[d] {
			continue
		}
		seen[d] = true
		parts = append(parts, d)
	}
	return strings.Join(parts, string(os.PathListSeparator))
}
