// Package preset holds hardcoded bundle data for the MVP UX prototype.
// In v0.1+ this will fetch TOML from raw.githubusercontent.com.
package preset

// Tool represents a single installable item.
//
// InstallMac and InstallLinux capture the platform-specific install
// command(s). When the real installer pass lands these will be the
// shell commands actually run (after detection / dry-run preview).
// Sourced from each tool's official docs (verified via perplexity for
// AI CLIs).
type Tool struct {
	Name         string // display name
	Description  string // short blurb
	Source       string // brew / cask / apt / mise / npm / curl / custom
	Installed    bool   // populated by detect pass; hardcoded for UX demo
	Version      string // e.g. "2.42.0" when Installed == true
	InstallMac   string // shell command for macOS
	InstallLinux string // shell command for Debian/Ubuntu Linux
	Binary       string // binary name to detect (defaults to Name if empty)
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
				{Name: "git", Source: "brew", Installed: true, Version: "2.42.0",
					InstallMac: "brew install git", InstallLinux: "sudo apt-get install -y git"},
				{Name: "gh", Source: "brew", Installed: true, Version: "2.54.0",
					InstallMac: "brew install gh", InstallLinux: "sudo apt-get install -y gh"},
				{Name: "fzf", Source: "brew", Installed: true, Version: "0.55.0",
					InstallMac: "brew install fzf", InstallLinux: "sudo apt-get install -y fzf"},
				{Name: "ripgrep", Source: "brew", Binary: "rg",
					InstallMac: "brew install ripgrep", InstallLinux: "sudo apt-get install -y ripgrep"},
				{Name: "bat", Source: "brew", Installed: true, Version: "0.24.0",
					InstallMac: "brew install bat", InstallLinux: "sudo apt-get install -y bat"},
				{Name: "eza", Source: "brew", Installed: true, Version: "0.18.0",
					InstallMac: "brew install eza", InstallLinux: "sudo apt-get install -y eza"},
				{Name: "fd", Source: "brew", Installed: true, Version: "9.0.0",
					InstallMac: "brew install fd", InstallLinux: "sudo apt-get install -y fd-find"},
				{Name: "zoxide", Source: "brew", Installed: true, Version: "0.9.4",
					InstallMac: "brew install zoxide", InstallLinux: "sudo apt-get install -y zoxide"},
				{Name: "jq", Source: "brew", Installed: true, Version: "1.7.1",
					InstallMac: "brew install jq", InstallLinux: "sudo apt-get install -y jq"},
				{Name: "tree", Source: "brew",
					InstallMac: "brew install tree", InstallLinux: "sudo apt-get install -y tree"},
				{Name: "tmux", Source: "brew", Installed: true, Version: "3.4",
					InstallMac: "brew install tmux", InstallLinux: "sudo apt-get install -y tmux"},
				{Name: "lazygit", Source: "brew", Installed: true, Version: "0.43.0",
					InstallMac:   "brew install lazygit",
					InstallLinux: "sudo add-apt-repository -y ppa:lazygit-team/release && sudo apt-get install -y lazygit"},
				{Name: "btop", Source: "brew", Installed: true, Version: "1.3.2",
					InstallMac: "brew install btop", InstallLinux: "sudo apt-get install -y btop"},
				{Name: "glow", Source: "brew", Installed: true, Version: "2.0.0",
					InstallMac: "brew install glow", InstallLinux: "sudo apt-get install -y glow"},
				{Name: "yazi", Source: "brew", Installed: true, Version: "0.3.0",
					InstallMac: "brew install yazi", InstallLinux: "cargo install --locked yazi-fm yazi-cli"},
				{Name: "mise", Source: "brew", Installed: true, Version: "2024.9",
					InstallMac: "brew install mise", InstallLinux: "curl -fsSL https://mise.run | sh"},
				{Name: "starship", Source: "custom",
					InstallMac:   "curl -sS https://starship.rs/install.sh | sh -s -- -y",
					InstallLinux: "curl -sS https://starship.rs/install.sh | sh -s -- -y"},
				{Name: "node (lts)", Source: "mise", Binary: "node",
					InstallMac: "mise use -g node@lts", InstallLinux: "mise use -g node@lts"},
				{Name: "bun", Source: "mise",
					InstallMac: "mise use -g bun@latest", InstallLinux: "mise use -g bun@latest"},
				{Name: "pnpm", Source: "mise",
					InstallMac: "mise use -g pnpm@latest", InstallLinux: "mise use -g pnpm@latest"},
				{Name: "uv", Source: "mise",
					InstallMac: "mise use -g uv@latest", InstallLinux: "mise use -g uv@latest"},
				{Name: "go", Source: "mise", Installed: true, Version: "1.26.0",
					InstallMac: "mise use -g go@latest", InstallLinux: "mise use -g go@latest"},
			},
		},
		{
			ID:          "mac-power-user",
			Name:        "mac-power-user",
			Description: "Mac CLI utilities (blueutil, brightness, dockutil, ...)",
			Tools: []Tool{
				{Name: "blueutil", Source: "brew", Installed: true, Version: "2.11.0",
					InstallMac: "brew install blueutil"},
				{Name: "brightness", Source: "brew", Installed: true, Version: "1.2",
					InstallMac: "brew install brightness"},
				{Name: "dockutil", Source: "brew", Installed: true, Version: "3.1.3",
					InstallMac: "brew install dockutil"},
				{Name: "switchaudio-osx", Source: "brew", Installed: true, Version: "1.2.4",
					InstallMac: "brew install switchaudio-osx"},
				{Name: "mas", Source: "brew", InstallMac: "brew install mas"},
				{Name: "raycast", Source: "cask", Installed: true, Version: "cask",
					InstallMac: "brew install --cask raycast"},
				{Name: "maccy", Source: "cask", Installed: true, Version: "cask",
					InstallMac: "brew install --cask maccy"},
				{Name: "keycastr", Source: "cask",
					InstallMac: "brew install --cask keycastr"},
				{Name: "monitorcontrol", Source: "cask", Installed: true, Version: "cask",
					InstallMac: "brew install --cask monitorcontrol"},
			},
		},
		{
			// AI CLIs — three picks (codex, claude code, opencode).
			// All installed via npm globally for cross-platform consistency.
			// Requires Node.js (≥18 for claude-code, ≥22 recommended for codex).
			ID:          "ai-clis",
			Name:        "ai-clis",
			Description: "codex, claude-code, opencode (npm)",
			Tools: []Tool{
				{
					Name: "codex", Source: "npm",
					Description: "OpenAI Codex CLI (@openai/codex)",
					Installed:   true, Version: "0.14.0",
					InstallMac:   "npm install -g @openai/codex",
					InstallLinux: "npm install -g @openai/codex",
					Binary:       "codex",
				},
				{
					Name: "claude-code", Source: "npm",
					Description: "Anthropic Claude Code CLI (@anthropic-ai/claude-code)",
					Installed:   true, Version: "2.1.123",
					InstallMac:   "npm install -g @anthropic-ai/claude-code",
					InstallLinux: "npm install -g @anthropic-ai/claude-code",
					Binary:       "claude",
				},
				{
					Name: "opencode", Source: "npm",
					Description: "Open-source TUI coding agent (opencode-ai)",
					InstallMac:   "npm install -g opencode-ai",
					InstallLinux: "npm install -g opencode-ai",
					Binary:       "opencode",
				},
			},
		},
		{
			ID:          "media",
			Name:        "media",
			Description: "ffmpeg, yt-dlp, imagemagick, etc.",
			Tools: []Tool{
				{Name: "ffmpeg", Source: "brew", Installed: true, Version: "7.0",
					InstallMac: "brew install ffmpeg", InstallLinux: "sudo apt-get install -y ffmpeg"},
				{Name: "yt-dlp", Source: "brew", Installed: true, Version: "2024.9",
					InstallMac: "brew install yt-dlp", InstallLinux: "sudo apt-get install -y yt-dlp"},
				{Name: "imagemagick", Source: "brew", Installed: true, Version: "7.1.1",
					InstallMac: "brew install imagemagick", InstallLinux: "sudo apt-get install -y imagemagick"},
				{Name: "poppler", Source: "brew", Installed: true, Version: "24.08",
					InstallMac: "brew install poppler", InstallLinux: "sudo apt-get install -y poppler-utils"},
				{Name: "exiftool", Source: "brew",
					InstallMac: "brew install exiftool", InstallLinux: "sudo apt-get install -y libimage-exiftool-perl"},
			},
		},
	}
}
