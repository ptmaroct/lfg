// Package preset holds hardcoded bundle data for the MVP UX prototype.
// In v0.1+ this will fetch TOML from raw.githubusercontent.com.
package preset

// Tool represents a single installable item.
type Tool struct {
	Name        string // display name
	Description string // short blurb
	Source      string // brew / apt / mise / custom
	Installed   bool   // populated by detect pass; hardcoded for UX demo
	Version     string // e.g. "2.42.0" when Installed == true
}

// Bundle is a named group of tools the user can toggle on/off.
type Bundle struct {
	ID          string
	Name        string
	Description string
	Default     bool
	Tools       []Tool
}

// All returns the bundles shipped with the CLI. Hardcoded for prototype.
func All() []Bundle {
	return []Bundle{
		{
			ID:          "default",
			Name:        "default",
			Description: "Universal dev starter (~20 tools)",
			Default:     true,
			Tools: []Tool{
				{Name: "git", Source: "brew", Installed: true, Version: "2.42.0"},
				{Name: "gh", Source: "brew", Installed: true, Version: "2.54.0"},
				{Name: "fzf", Source: "brew", Installed: true, Version: "0.55.0"},
				{Name: "ripgrep", Source: "brew"},
				{Name: "bat", Source: "brew", Installed: true, Version: "0.24.0"},
				{Name: "eza", Source: "brew", Installed: true, Version: "0.18.0"},
				{Name: "fd", Source: "brew", Installed: true, Version: "9.0.0"},
				{Name: "zoxide", Source: "brew", Installed: true, Version: "0.9.4"},
				{Name: "jq", Source: "brew", Installed: true, Version: "1.7.1"},
				{Name: "tree", Source: "brew"},
				{Name: "tmux", Source: "brew", Installed: true, Version: "3.4"},
				{Name: "lazygit", Source: "brew", Installed: true, Version: "0.43.0"},
				{Name: "btop", Source: "brew", Installed: true, Version: "1.3.2"},
				{Name: "glow", Source: "brew", Installed: true, Version: "2.0.0"},
				{Name: "yazi", Source: "brew", Installed: true, Version: "0.3.0"},
				{Name: "mise", Source: "brew", Installed: true, Version: "2024.9"},
				{Name: "starship", Source: "custom"},
				{Name: "node (lts)", Source: "mise"},
				{Name: "bun", Source: "mise"},
				{Name: "pnpm", Source: "mise"},
				{Name: "uv", Source: "mise"},
				{Name: "go", Source: "mise", Installed: true, Version: "1.26.0"},
			},
		},
		{
			ID:          "mac-power-user",
			Name:        "mac-power-user",
			Description: "Mac CLI utilities (blueutil, brightness, dockutil, ...)",
			Tools: []Tool{
				{Name: "blueutil", Source: "brew", Installed: true, Version: "2.11.0"},
				{Name: "brightness", Source: "brew", Installed: true, Version: "1.2"},
				{Name: "dockutil", Source: "brew", Installed: true, Version: "3.1.3"},
				{Name: "switchaudio-osx", Source: "brew", Installed: true, Version: "1.2.4"},
				{Name: "mas", Source: "brew"},
				{Name: "raycast", Source: "cask", Installed: true, Version: "cask"},
				{Name: "maccy", Source: "cask", Installed: true, Version: "cask"},
				{Name: "keycastr", Source: "cask"},
				{Name: "monitorcontrol", Source: "cask", Installed: true, Version: "cask"},
			},
		},
		{
			ID:          "ai-clis",
			Name:        "ai-clis",
			Description: "codex, claude, gemini, crush and friends",
			Tools: []Tool{
				{Name: "@openai/codex", Source: "npm", Installed: true, Version: "0.14.0"},
				{Name: "gemini-cli", Source: "brew", Installed: true, Version: "0.2.0"},
				{Name: "crush", Source: "brew", Installed: true, Version: "0.3.0"},
				{Name: "happy-coder", Source: "npm", Installed: true, Version: "1.5.0"},
				{Name: "agent-browser", Source: "brew", Installed: true, Version: "0.4.0"},
				{Name: "block-goose-cli", Source: "brew", Installed: true, Version: "0.2.0"},
			},
		},
		{
			ID:          "media",
			Name:        "media",
			Description: "ffmpeg, yt-dlp, imagemagick, etc.",
			Tools: []Tool{
				{Name: "ffmpeg", Source: "brew", Installed: true, Version: "7.0"},
				{Name: "yt-dlp", Source: "brew", Installed: true, Version: "2024.9"},
				{Name: "imagemagick", Source: "brew", Installed: true, Version: "7.1.1"},
				{Name: "poppler", Source: "brew", Installed: true, Version: "24.08"},
				{Name: "exiftool", Source: "brew"},
			},
		},
	}
}
