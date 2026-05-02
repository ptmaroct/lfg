// Package preset holds the bundle/tool data model and the built-in
// preset shipped with lfg. v0.1+ supports loading custom presets via
// preset.Load — see loader.go.
package preset

// Tool represents a single installable item.
//
// InstallMac and InstallLinux capture the platform-specific install
// command(s). Detect populates Installed/Version at runtime based on
// what's actually present on the host.
type Tool struct {
	Name         string `toml:"name"`
	Description  string `toml:"description,omitempty"`
	Source       string `toml:"source"` // brew / cask / apt / mise / npm / curl / custom / skills
	Installed    bool   `toml:"-"`      // populated by detect pass; never serialized
	Version      string `toml:"-"`      // populated by detect pass; never serialized
	InstallMac   string `toml:"install_mac,omitempty"`
	InstallLinux string `toml:"install_linux,omitempty"`
	Binary       string `toml:"binary,omitempty"`    // binary on PATH; defaults to Name
	SkillURL     string `toml:"skill_url,omitempty"` // for Source="skills"
	// Mandatory rows can't be unselected in the tree picker. Used for
	// hard dependencies like Homebrew that other tools build on.
	Mandatory bool `toml:"mandatory,omitempty"`
}

// Bundle is a named group of tools the user can toggle on/off.
type Bundle struct {
	ID          string `toml:"id"`
	Name        string `toml:"name"`
	Description string `toml:"description,omitempty"`
	Default     bool   `toml:"default,omitempty"`
	Tools       []Tool `toml:"tools"`
}

// All returns the bundles shipped with the CLI by default. Three
// bundles: barebones (universal foundation), dev-tools (AI coding
// CLIs), skills (cross-harness skill packs).
func All() []Bundle {
	return []Bundle{
		{
			ID:          "barebones",
			Name:        "barebones",
			Description: "Universal foundation: Homebrew, Node, Python, runtime managers",
			Default:     true,
			Tools: []Tool{
				// Homebrew — mandatory. Other tools install via brew.
				{
					Name: "brew", Source: "custom", Binary: "brew", Mandatory: true,
					Description:  "Homebrew package manager — required by everything below",
					InstallMac:   `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`,
					InstallLinux: `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`,
				},
				// Runtime managers + language runtimes.
				{
					Name: "mise", Source: "brew", Binary: "mise",
					Description:  "Polyglot runtime manager (replaces nvm/pyenv/rbenv)",
					InstallMac:   "brew install mise",
					InstallLinux: "curl -fsSL https://mise.run | sh",
				},
				{
					Name: "node-lts", Source: "mise", Binary: "node",
					Description:  "Node.js LTS via mise",
					InstallMac:   "mise use -g node@lts",
					InstallLinux: "mise use -g node@lts",
				},
				{
					Name: "pnpm", Source: "mise", Binary: "pnpm",
					InstallMac:   "mise use -g pnpm@latest",
					InstallLinux: "mise use -g pnpm@latest",
				},
				{
					Name: "bun", Source: "mise", Binary: "bun",
					InstallMac:   "mise use -g bun@latest",
					InstallLinux: "mise use -g bun@latest",
				},
				{
					Name: "yarn", Source: "npm", Binary: "yarn",
					Description:  "Yarn classic via corepack-style global install",
					InstallMac:   "npm install -g yarn",
					InstallLinux: "npm install -g yarn",
				},
				{
					Name: "python", Source: "mise", Binary: "python",
					Description:  "Python latest via mise",
					InstallMac:   "mise use -g python@latest",
					InstallLinux: "mise use -g python@latest",
				},
				{
					Name: "uv", Source: "custom", Binary: "uv",
					Description:  "Astral uv — fast Python package manager",
					InstallMac:   "curl -LsSf https://astral.sh/uv/install.sh | sh",
					InstallLinux: "curl -LsSf https://astral.sh/uv/install.sh | sh",
				},
			},
		},
		{
			ID:          "dev-tools",
			Name:        "dev-tools",
			Description: "AI coding CLIs (Claude Code, Codex, OpenCode, Droid)",
			Tools: []Tool{
				{
					Name: "claude-code", Source: "custom", Binary: "claude",
					Description:  "Anthropic Claude Code CLI",
					InstallMac:   "curl -fsSL https://claude.ai/install.sh | bash",
					InstallLinux: "curl -fsSL https://claude.ai/install.sh | bash",
				},
				{
					Name: "codex", Source: "npm", Binary: "codex",
					Description:  "OpenAI Codex CLI (@openai/codex)",
					InstallMac:   "npm install -g @openai/codex",
					InstallLinux: "npm install -g @openai/codex",
				},
				{
					Name: "opencode", Source: "custom", Binary: "opencode",
					Description:  "sst/opencode — open-source coding agent",
					InstallMac:   "curl -fsSL https://opencode.ai/install | bash",
					InstallLinux: "curl -fsSL https://opencode.ai/install | bash",
				},
				{
					Name: "droid", Source: "custom", Binary: "droid",
					Description:  "Factory AI Droid CLI",
					InstallMac:   "curl -fsSL https://app.factory.ai/cli | sh",
					InstallLinux: "curl -fsSL https://app.factory.ai/cli | sh",
				},
			},
		},
		{
			ID:          "skills",
			Name:        "skills",
			Description: "Agent skills installed via `npx skills add` (cross-harness)",
			Tools: []Tool{
				{
					Name: "agent-browser", Source: "skills",
					Description: "Headless-browser agent skill",
					SkillURL:    "https://github.com/anthropics/skills/tree/main/agent-browser",
				},
				{
					Name: "frontend-design", Source: "skills",
					Description: "UI/design review skill",
					SkillURL:    "https://github.com/anthropics/skills/tree/main/frontend-design",
				},
				{
					Name: "db-introspect", Source: "skills",
					Description: "Database schema introspection skill",
					SkillURL:    "https://github.com/anthropics/skills/tree/main/db-introspect",
				},
				{
					Name: "prompt-eval", Source: "skills",
					Description: "Prompt eval harness skill",
					SkillURL:    "https://github.com/anthropics/skills/tree/main/prompt-eval",
				},
				{
					Name: "portless", Source: "skills",
					Description: "vercel-labs/portless tunnel skill",
					SkillURL:    "https://github.com/vercel-labs/portless",
				},
			},
		},
	}
}
