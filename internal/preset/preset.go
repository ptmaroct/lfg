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
	// Homepage is the project's website / repo URL. Surfaced in the
	// info dialog so users can read more before installing.
	Homepage string `toml:"homepage,omitempty"`
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
				{
					Name: "brew", Source: "curl", Binary: "brew", Mandatory: true,
					Description:  "Homebrew package manager — required by everything below",
					Homepage:     "https://brew.sh",
					InstallMac:   `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`,
					InstallLinux: `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`,
				},
				{
					Name: "mise", Source: "brew", Binary: "mise",
					Description:  "Polyglot runtime manager (replaces nvm/pyenv/rbenv)",
					Homepage:     "https://mise.jdx.dev",
					InstallMac:   "brew install mise",
					InstallLinux: "curl -fsSL https://mise.run | sh",
				},
				{
					Name: "node-lts", Source: "mise", Binary: "node",
					Description:  "Node.js LTS via mise",
					Homepage:     "https://nodejs.org",
					InstallMac:   "mise use -g node@lts",
					InstallLinux: "mise use -g node@lts",
				},
				{
					Name: "pnpm", Source: "mise", Binary: "pnpm",
					Description:  "Fast, disk-space-efficient package manager",
					Homepage:     "https://pnpm.io",
					InstallMac:   "mise use -g pnpm@latest",
					InstallLinux: "mise use -g pnpm@latest",
				},
				{
					Name: "bun", Source: "mise", Binary: "bun",
					Description:  "All-in-one JavaScript runtime, bundler, package manager",
					Homepage:     "https://bun.sh",
					InstallMac:   "mise use -g bun@latest",
					InstallLinux: "mise use -g bun@latest",
				},
				{
					Name: "yarn", Source: "npm", Binary: "yarn",
					Description:  "Yarn classic — JavaScript package manager",
					Homepage:     "https://yarnpkg.com",
					InstallMac:   "npm install -g yarn",
					InstallLinux: "npm install -g yarn",
				},
				{
					Name: "python", Source: "mise", Binary: "python",
					Description:  "Python latest via mise",
					Homepage:     "https://python.org",
					InstallMac:   "mise use -g python@latest",
					InstallLinux: "mise use -g python@latest",
				},
				{
					Name: "uv", Source: "curl", Binary: "uv",
					Description:  "Astral uv — fast Python package + project manager",
					Homepage:     "https://docs.astral.sh/uv",
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
					Name: "claude-code", Source: "curl", Binary: "claude",
					Description:  "Anthropic's Claude Code CLI — agentic coding in your terminal",
					Homepage:     "https://docs.claude.com/en/docs/claude-code",
					InstallMac:   "curl -fsSL https://claude.ai/install.sh | bash",
					InstallLinux: "curl -fsSL https://claude.ai/install.sh | bash",
				},
				{
					Name: "codex", Source: "npm", Binary: "codex",
					Description:  "OpenAI Codex CLI — coding agent from the makers of ChatGPT",
					Homepage:     "https://github.com/openai/codex",
					InstallMac:   "npm install -g @openai/codex",
					InstallLinux: "npm install -g @openai/codex",
				},
				{
					Name: "opencode", Source: "curl", Binary: "opencode",
					Description:  "sst/opencode — open-source TUI coding agent",
					Homepage:     "https://opencode.ai",
					InstallMac:   "curl -fsSL https://opencode.ai/install | bash",
					InstallLinux: "curl -fsSL https://opencode.ai/install | bash",
				},
				{
					Name: "droid", Source: "curl", Binary: "droid",
					Description:  "Factory AI Droid CLI — software-engineering agents in your terminal",
					Homepage:     "https://docs.factory.ai/cli/getting-started/quickstart",
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
					Description: "vercel-labs/agent-browser — headless-browser agent skill",
					Homepage:    "https://github.com/vercel-labs/agent-browser",
					SkillURL:    "https://github.com/vercel-labs/agent-browser",
				},
				{
					Name: "frontend-design", Source: "skills",
					Description: "UI/design review skill",
					Homepage:    "https://github.com/anthropics/skills/tree/main/frontend-design",
					SkillURL:    "https://github.com/anthropics/skills/tree/main/frontend-design",
				},
				{
					Name: "db-introspect", Source: "skills",
					Description: "Database schema introspection skill",
					Homepage:    "https://github.com/anthropics/skills/tree/main/db-introspect",
					SkillURL:    "https://github.com/anthropics/skills/tree/main/db-introspect",
				},
				{
					Name: "prompt-eval", Source: "skills",
					Description: "Prompt eval harness skill",
					Homepage:    "https://github.com/anthropics/skills/tree/main/prompt-eval",
					SkillURL:    "https://github.com/anthropics/skills/tree/main/prompt-eval",
				},
				{
					Name: "portless", Source: "skills",
					Description: "vercel-labs/portless — local hostnames for dev servers",
					Homepage:    "https://github.com/vercel-labs/portless",
					SkillURL:    "https://github.com/vercel-labs/portless",
				},
			},
		},
	}
}
