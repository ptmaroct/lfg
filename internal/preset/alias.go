package preset

import "runtime"

// Alias is a single shell alias the user can opt into during setup.
// Written into a fenced lfg-managed block in every detected shell rc
// by installer.EnsureShellAliases. Top-level (not nested in Bundle)
// because alias selection is orthogonal to the install plan: a user
// without claude installed can still want `c` / `cr` configured for
// later, and aliases survive a `chsh`.
type Alias struct {
	Name        string `toml:"name"`
	Command     string `toml:"command"`
	FishCommand string `toml:"fish_command,omitempty"`
	Description string `toml:"description,omitempty"`
	Group       string `toml:"group,omitempty"`
	Default     bool   `toml:"default,omitempty"`
	SkipMac     bool   `toml:"skip_mac,omitempty"`
	SkipLinux   bool   `toml:"skip_linux,omitempty"`
}

// AliasGroup buckets aliases by intent for the picker UI.
type AliasGroup struct {
	ID          string
	Name        string
	Description string
	Aliases     []Alias
}

const (
	AliasGroupGit    = "git"
	AliasGroupClaude = "claude"
	AliasGroupShell  = "shell"
)

// DefaultAliasesFlat is the hardcoded default catalog (used when no
// `--config` provides its own [[aliases]] table).
func DefaultAliasesFlat() []Alias {
	return []Alias{
		{Name: "gd", Command: "git checkout develop && git pull", Description: "checkout develop + pull", Group: AliasGroupGit, Default: true},
		{Name: "gf", Command: "git fetch --all --prune", Description: "fetch all remotes, prune", Group: AliasGroupGit, Default: true},
		{Name: "gs", Command: "git status", Description: "git status", Group: AliasGroupGit, Default: true},
		{Name: "gp", Command: "git push", Description: "git push", Group: AliasGroupGit, Default: true},
		{Name: "gl", Command: "git log --oneline -20", Description: "compact recent log", Group: AliasGroupGit, Default: true},
		{Name: "gco", Command: "git checkout", Description: "git checkout", Group: AliasGroupGit, Default: true},

		{Name: "c", Command: "claude", Description: "claude (normal mode)", Group: AliasGroupClaude, Default: true},
		{Name: "cr", Command: "claude --resume", Description: "claude resume", Group: AliasGroupClaude, Default: true},
		{Name: "cw", Command: "claude worktree", Description: "claude worktree", Group: AliasGroupClaude, Default: true},

		// reload is per-shell special-cased by buildAliasBlock — bash/zsh
		// emit `alias reload='exec bash|zsh'`, fish emits a function.
		{Name: "reload", Command: "exec $SHELL", FishCommand: "exec fish", Description: "reload current shell", Group: AliasGroupShell, Default: true},
	}
}

// DefaultAliases returns the default catalog grouped for picker UI.
func DefaultAliases() []AliasGroup {
	return GroupAliases(DefaultAliasesFlat())
}

// GroupAliases buckets a flat slice into ordered AliasGroups.
// Groups appear in the canonical order: git, claude, shell, then any
// custom Group values in first-seen order. Aliases keep their input
// order within each group.
func GroupAliases(flat []Alias) []AliasGroup {
	if len(flat) == 0 {
		return nil
	}
	meta := map[string]struct{ name, desc string }{
		AliasGroupGit:    {"git", "git workflow shortcuts"},
		AliasGroupClaude: {"claude", "claude code shortcuts"},
		AliasGroupShell:  {"shell", "shell helpers"},
	}
	order := []string{AliasGroupGit, AliasGroupClaude, AliasGroupShell}
	seen := map[string]bool{AliasGroupGit: true, AliasGroupClaude: true, AliasGroupShell: true}
	bucket := map[string][]Alias{}
	for _, a := range flat {
		g := a.Group
		if g == "" {
			g = "custom"
		}
		if !seen[g] {
			seen[g] = true
			order = append(order, g)
		}
		bucket[g] = append(bucket[g], a)
	}
	out := make([]AliasGroup, 0, len(order))
	for _, id := range order {
		if len(bucket[id]) == 0 {
			continue
		}
		m, ok := meta[id]
		if !ok {
			m = struct{ name, desc string }{id, ""}
		}
		out = append(out, AliasGroup{ID: id, Name: m.name, Description: m.desc, Aliases: bucket[id]})
	}
	return out
}

// FilterAliasesForHost drops aliases whose Skip{Mac,Linux} matches
// the running OS. Mirror of FilterForHost (preset.go:12).
func FilterAliasesForHost(in []Alias) []Alias {
	out := make([]Alias, 0, len(in))
	for _, a := range in {
		if runtime.GOOS == "darwin" && a.SkipMac {
			continue
		}
		if runtime.GOOS != "darwin" && a.SkipLinux {
			continue
		}
		out = append(out, a)
	}
	return out
}
