package backup

import (
	"os"
	"path/filepath"
)

// Source is one item we attempt to capture. Missing items are skipped
// silently — the whitelist is intentionally optimistic.
type Source struct {
	// Path is absolute (resolved from $HOME).
	Path string
	// ArchiveName is the path stored inside the tar; defaults to a
	// home-relative version of Path.
	ArchiveName string
	// Optional makes the source's absence not even worth logging.
	Optional bool
}

// homeDir is overridable in tests. Real callers use os.UserHomeDir.
var homeDir = func() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return h
}

// DefaultSources returns the curated list (read-only). Exposed so the
// TUI can preview what's about to be backed up before the user
// confirms the lock-it-or-not prompt.
func DefaultSources() []Source { return defaultSources() }

// defaultSources returns the curated list of dotfiles + config dirs
// captured by `lfg backup` for v0.1. Order matters only for stable
// archive output. Items rooted in $HOME are stored relative to it.
func defaultSources() []Source {
	home := homeDir()
	if home == "" {
		return nil
	}
	rel := func(rel string) Source {
		return Source{Path: filepath.Join(home, rel), ArchiveName: rel, Optional: true}
	}
	return []Source{
		// Shell rc files
		rel(".zshrc"),
		rel(".zprofile"),
		rel(".zshenv"),
		rel(".bashrc"),
		rel(".bash_profile"),
		rel(".profile"),

		// Top-level dotfiles
		rel(".gitconfig"),
		rel(".tmux.conf"),
		rel(".vimrc"),
		rel(".editorconfig"),
		rel(".inputrc"),

		// ~/.config/* whitelist — pulled wholesale
		rel(".config/starship.toml"),
		rel(".config/starship"),
		rel(".config/mise"),
		rel(".config/bat"),
		rel(".config/btop"),
		rel(".config/lazygit"),
		rel(".config/yazi"),
		rel(".config/glow"),
		rel(".config/ghostty"),
		rel(".config/nvim"),
		rel(".config/zed"),
		rel(".config/zellij"),

		// AI tooling configs (user-level only, no project-local)
		rel(".claude/settings.json"),
		rel(".claude/CLAUDE.md"),
		rel(".claude/agents"),
		rel(".claude/commands"),
		rel(".codex/config.toml"),

		// SSH — capture the whole dir; private keys (id_*) are filtered
		// out by name pattern unless --include-ssh-keys + --encrypt.
		rel(".ssh"),
	}
}
