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
	Source       string `toml:"source"` // brew / cask / apt / mise / npm / curl / custom / skills / mcp
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
	// PostInstall lists shell commands to run sequentially after the
	// main install step succeeds. Useful when a tool needs runtime
	// prerequisites the primary installer doesn't cover (e.g.
	// agent-browser ships as a skill stub but needs `npm i -g
	// agent-browser` + `agent-browser install` to actually function).
	// Each line runs through the same exec path as the main install
	// so output streams into the live log + transcript file.
	PostInstall []string `toml:"post_install,omitempty"`
	// MCPType selects the transport for Source="mcp" entries:
	// "stdio" (default, local process — needs MCPPackage), "http", or
	// "sse" (both remote — need MCPURL). Empty string is treated as
	// "stdio" for backward compat.
	MCPType string `toml:"mcp_type,omitempty"`
	// MCPPackage is the npm package providing a stdio MCP server
	// (e.g. "@modelcontextprotocol/server-github"). Installed globally
	// via `npm install -g <pkg>`. Required for stdio; ignored for
	// http/sse.
	MCPPackage string `toml:"mcp_package,omitempty"`
	// MCPCommand + MCPArgs override the launch command for stdio MCPs
	// when the published binary name doesn't match the npm package
	// (default command = `npx -y <MCPPackage>`). Rarely needed.
	MCPCommand string   `toml:"mcp_command,omitempty"`
	MCPArgs    []string `toml:"mcp_args,omitempty"`
	// MCPURL is the endpoint for remote (http / sse) MCP servers.
	// Required when MCPType is "http" or "sse"; ignored for stdio.
	MCPURL string `toml:"mcp_url,omitempty"`
	// MCPHeaders is the optional set of HTTP headers sent with every
	// request to a remote MCP server (e.g. {"Authorization": "Bearer ..."}).
	// Values can reference env vars via $VAR — harnesses expand at
	// connect time.
	MCPHeaders map[string]string `toml:"mcp_headers,omitempty"`
	// TargetHarnesses lists the AI-coding-agent configs the MCP server
	// should be auto-registered with (e.g. ["claude-code", "codex"]).
	// Empty list = install only, no harness wiring.
	TargetHarnesses []string `toml:"target_harnesses,omitempty"`
	// EnvVars names the environment variables the tool needs at runtime
	// (e.g. ["GITHUB_PERSONAL_ACCESS_TOKEN"]). Surfaced in the info
	// dialog before install and on the done screen after, so the user
	// knows exactly what to set. The credentials wizard (between
	// confirm and progress) prompts password-style inputs for each one;
	// values land in MCPCredentials for the installer to use, and any
	// var the user leaves blank is shown on the done screen as a
	// reminder to set it manually.
	EnvVars []string `toml:"env_vars,omitempty"`
	// MCPCredentials maps EnvVar name → user-provided value, populated
	// at runtime by the credentials wizard. Never serialized — secrets
	// only live in memory + in the harness configs the installer writes
	// (which lfg treats as the user's responsibility to protect).
	MCPCredentials map[string]string `toml:"-"`
}

// PlannedVersion returns the version string lfg will install for this
// tool, parsed from its install command. Three patterns:
//
//   - mise/npm style ending in `@<ver>` → returns the literal token
//     ("lts", "latest", "20.11.0", etc.)
//   - known package-manager sources without an explicit @ver → "latest"
//     (brew, mise, npm, cask)
//   - everything else (curl scripts, skill stubs, custom) → "" so the
//     caller can fall back to "—"
//
// Cheap and parser-only — no network. Trades exactness for instant
// render; an authoritative version would need a registry lookup per
// tool which would freeze the tree picker on every open.
func (t Tool) PlannedVersion() string {
	cmd := t.InstallMac
	if cmd == "" {
		cmd = t.InstallLinux
	}
	if cmd != "" {
		// Match the LAST `@token` not preceded by `/` so we skip npm
		// scoped packages like `@openai/codex` and only catch the
		// version pin (`node@lts`, `pnpm@latest`, `node@20.11.0`).
		// Right-to-left scan keeps the regex simple.
		fields := splitCmd(cmd)
		for i := len(fields) - 1; i >= 0; i-- {
			f := fields[i]
			if at := lastAt(f); at > 0 && f[at-1] != '/' {
				v := f[at+1:]
				v = trimTrailingPunct(v)
				if v != "" {
					return v
				}
			}
		}
	}
	switch t.Source {
	case "brew", "cask", "mise", "npm":
		return "latest"
	}
	return ""
}

// splitCmd splits on whitespace + pipe so a piped install (`curl ... |
// sh`) is tokenized. Avoids importing strings here to keep the file
// self-contained — adapted-Fields.
func splitCmd(s string) []string {
	out := []string{}
	cur := []byte{}
	flush := func() {
		if len(cur) > 0 {
			out = append(out, string(cur))
			cur = cur[:0]
		}
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' || c == '\t' || c == '|' || c == '\n' {
			flush()
			continue
		}
		cur = append(cur, c)
	}
	flush()
	return out
}

func lastAt(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '@' {
			return i
		}
	}
	return -1
}

func trimTrailingPunct(s string) string {
	for len(s) > 0 {
		c := s[len(s)-1]
		if c == '"' || c == '\'' || c == ',' || c == ')' || c == '}' {
			s = s[:len(s)-1]
			continue
		}
		break
	}
	return s
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
					// The skill stub from `npx skills add` is just a
					// SKILL.md pointer; the actual `agent-browser`
					// binary + Chrome download are installed by the
					// two commands below. Without them the skill loads
					// but `agent-browser --version` fails with command
					// not found.
					PostInstall: []string{
						"npm install -g agent-browser",
						"agent-browser install",
					},
				},
				{
					Name: "frontend-design", Source: "skills",
					Description: "UI/design review skill",
					Homepage:    "https://github.com/anthropics/skills/tree/main/skills/frontend-design",
					SkillURL:    "https://github.com/anthropics/skills/tree/main/skills/frontend-design",
				},
				{
					Name: "claude-api", Source: "skills",
					Description: "Build apps on the Anthropic SDK with prompt caching",
					Homepage:    "https://github.com/anthropics/skills/tree/main/skills/claude-api",
					SkillURL:    "https://github.com/anthropics/skills/tree/main/skills/claude-api",
				},
				{
					Name: "skill-creator", Source: "skills",
					Description: "Helper for authoring + evaluating new agent skills",
					Homepage:    "https://github.com/anthropics/skills/tree/main/skills/skill-creator",
					SkillURL:    "https://github.com/anthropics/skills/tree/main/skills/skill-creator",
				},
				{
					Name: "webapp-testing", Source: "skills",
					Description: "Exploratory web-app testing skill",
					Homepage:    "https://github.com/anthropics/skills/tree/main/skills/webapp-testing",
					SkillURL:    "https://github.com/anthropics/skills/tree/main/skills/webapp-testing",
				},
				{
					Name: "portless", Source: "skills",
					Description: "vercel-labs/portless — local hostnames for dev servers",
					Homepage:    "https://github.com/vercel-labs/portless",
					SkillURL:    "https://github.com/vercel-labs/portless",
				},
			},
		},
		{
			ID:          "mcp",
			Name:        "mcp",
			Description: "Official Model Context Protocol servers — extend any MCP-aware harness",
			Tools: []Tool{
				// --- Knowledge & search ---
				// Up-to-date library docs, fed straight into the model context.
				{
					Name: "context7", Source: "mcp", Binary: "context7",
					Description:     "Up-to-date library docs from context7.com (Upstash)",
					Homepage:        "https://context7.com",
					MCPType:         "http",
					MCPURL:          "https://mcp.context7.com/mcp",
					MCPHeaders:      map[string]string{"CONTEXT7_API_KEY": "${CONTEXT7_API_KEY}"},
					EnvVars:         []string{"CONTEXT7_API_KEY"},
					TargetHarnesses: []string{"all"},
				},
				// Perplexity Ask — web search with citations.
				{
					Name: "perplexity", Source: "mcp", Binary: "perplexity",
					Description:     "Perplexity Ask — web search w/ citations (official)",
					Homepage:        "https://docs.perplexity.ai/guides/mcp-server",
					MCPType:         "stdio",
					MCPPackage:      "@perplexity-ai/mcp-server",
					EnvVars:         []string{"PERPLEXITY_API_KEY"},
					TargetHarnesses: []string{"all"},
				},
				// Exa AI search — code, docs, web.
				{
					Name: "exa", Source: "mcp", Binary: "exa",
					Description:     "Exa AI search — code + docs + web (exa.ai official)",
					Homepage:        "https://github.com/exa-labs/exa-mcp-server",
					MCPType:         "stdio",
					MCPPackage:      "exa-mcp-server",
					EnvVars:         []string{"EXA_API_KEY"},
					TargetHarnesses: []string{"all"},
				},

				// --- Observability & DevOps ---
				// Sentry: errors, traces, releases. Hosted server uses OAuth on
				// first connect — no static token to collect.
				{
					Name: "sentry", Source: "mcp", Binary: "sentry",
					Description:     "Sentry errors + traces + releases (OAuth on first use)",
					Homepage:        "https://mcp.sentry.dev",
					MCPType:         "http",
					MCPURL:          "https://mcp.sentry.dev/mcp",
					TargetHarnesses: []string{"all"},
				},
				// CircleCI: builds, workflows, failed-job context.
				{
					Name: "circleci", Source: "mcp", Binary: "circleci",
					Description:     "CircleCI builds + workflows (CircleCI official)",
					Homepage:        "https://github.com/CircleCI-Public/mcp-server-circleci",
					MCPType:         "stdio",
					MCPPackage:      "@circleci/mcp-server-circleci",
					EnvVars:         []string{"CIRCLECI_TOKEN"},
					TargetHarnesses: []string{"all"},
				},

				// --- Productivity ---
				// Notion: pages, databases, search. Hosted server uses OAuth.
				{
					Name: "notion", Source: "mcp", Binary: "notion",
					Description:     "Notion pages + databases (OAuth on first use)",
					Homepage:        "https://developers.notion.com/docs/mcp",
					MCPType:         "http",
					MCPURL:          "https://mcp.notion.com/mcp",
					TargetHarnesses: []string{"all"},
				},
			},
		},
	}
}
