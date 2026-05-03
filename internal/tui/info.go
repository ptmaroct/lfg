package tui

import (
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ptmaroct/lfg/internal/preset"
)

// infoModel — read-only "what does this tool do" overlay shown when the
// user presses `i` on a row in the tree picker. Carries the previous
// screen so Esc/Enter goes back without losing the picker state.
type infoModel struct {
	palette  Palette
	tool     preset.Tool
	bundleID string
	prev     screen
}

func newInfo(p Palette, bundleID string, t preset.Tool, prev screen) infoModel {
	return infoModel{palette: p, tool: t, bundleID: bundleID, prev: prev}
}

func (m infoModel) Init() tea.Cmd { return nil }

func (m infoModel) Update(msg tea.Msg) (infoModel, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "esc", "enter", "i", "I", " ", "backspace", "delete":
			return m, func() tea.Msg { return closeInfoMsg{} }
		}
	}
	return m, nil
}

// closeInfoMsg returns to the screen the info dialog was opened from
// without rebuilding it (preserves cursor + selection state in the
// tree picker).
type closeInfoMsg struct{}

func (m infoModel) View(width, height int) string {
	p := m.palette
	canvasW := CanvasW(width)
	contentW := canvasW - 4

	label := lipgloss.NewStyle().Foreground(p.Muted).Bold(false)
	value := lipgloss.NewStyle().Foreground(p.Text)
	link := lipgloss.NewStyle().Foreground(p.Accent).Underline(true)

	row := func(k, v string) string {
		if v == "" {
			v = "—"
		}
		return "  " + label.Render(padRightPlain(k, 12)) + value.Render(v) + "\n"
	}
	linkRow := func(k, v string) string {
		if v == "" {
			v = "—"
		}
		return "  " + label.Render(padRightPlain(k, 12)) + link.Render(v) + "\n"
	}

	var b strings.Builder
	b.WriteString(SectionLabel(p, m.tool.Name, m.bundleID, contentW))
	b.WriteString("\n\n")

	if m.tool.Description != "" {
		b.WriteString("  " + lipgloss.NewStyle().Foreground(p.Text).Italic(true).
			Render(m.tool.Description) + "\n\n")
	}

	// Status line.
	status := "not installed"
	statusColor := p.Muted
	if m.tool.Installed {
		statusColor = p.Success
		if m.tool.Version != "" {
			status = "installed (v" + m.tool.Version + ")"
		} else {
			status = "installed"
		}
	}
	b.WriteString("  " + label.Render(padRightPlain("status", 12)) +
		lipgloss.NewStyle().Foreground(statusColor).Bold(true).Render(status) + "\n")

	b.WriteString(row("via", m.tool.Source))
	b.WriteString(row("binary", binaryName(m.tool)))
	b.WriteString(linkRow("homepage", m.tool.Homepage))
	if m.tool.SkillURL != "" {
		b.WriteString(linkRow("skill repo", m.tool.SkillURL))
	}
	if m.tool.MCPPackage != "" {
		b.WriteString(row("npm package", m.tool.MCPPackage))
	}
	if m.tool.MCPURL != "" {
		b.WriteString(row("mcp url", m.tool.MCPURL))
	}
	if m.tool.MCPType != "" {
		b.WriteString(row("transport", m.tool.MCPType))
	}
	if len(m.tool.TargetHarnesses) > 0 {
		harnesses := strings.Join(m.tool.TargetHarnesses, ", ")
		if len(m.tool.TargetHarnesses) == 1 && m.tool.TargetHarnesses[0] == "all" {
			harnesses = "all installed (claude-code, codex, opencode, droid)"
		}
		b.WriteString(row("harnesses", harnesses))
	}
	if m.tool.Mandatory {
		b.WriteString(row("mandatory", "yes — required by other tools"))
	}

	// Env vars section — for tools (typically MCP servers) that need
	// runtime secrets like GITHUB_PERSONAL_ACCESS_TOKEN. Surfaced here
	// so the user sees the requirement BEFORE installing, and again on
	// the done screen with copy-paste export lines.
	if len(m.tool.EnvVars) > 0 {
		b.WriteString("\n  " + label.Render("env vars (set before use)") + "\n")
		envStyle := lipgloss.NewStyle().Foreground(p.Warn).Bold(true)
		for _, ev := range m.tool.EnvVars {
			b.WriteString("    " + envStyle.Render(ev) + "\n")
		}
	}

	// Install command for current OS.
	cmd := m.tool.InstallMac
	if runtime.GOOS != "darwin" {
		cmd = m.tool.InstallLinux
	}
	if cmd != "" {
		b.WriteString("\n  " + label.Render("install command") + "\n")
		b.WriteString("  " + lipgloss.NewStyle().
			Foreground(p.Text).
			Background(p.Panel).
			Padding(0, 1).
			Render(cmd) + "\n")
	}

	// PostInstall: list any commands that run after the main install.
	// Useful for skills like agent-browser that need additional npm /
	// chrome bootstrapping before they're functional.
	if len(m.tool.PostInstall) > 0 {
		b.WriteString("\n  " + label.Render("post-install") + "\n")
		boxStyle := lipgloss.NewStyle().Foreground(p.Text).Background(p.Panel).Padding(0, 1)
		for _, c := range m.tool.PostInstall {
			b.WriteString("  " + boxStyle.Render(c) + "\n")
		}
	}

	return Frame(p, width, height,
		"info · "+m.tool.Name,
		b.String(),
		HintLine(p, KeyHint(p, "ESC", "back")),
		height < 22,
	)
}

func binaryName(t preset.Tool) string {
	if t.Binary != "" {
		return t.Binary
	}
	return t.Name
}
