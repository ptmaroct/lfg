package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ptmaroct/lfg/internal/preset"
)

// doneModel — final celebration card. Big checkmark, headline,
// next-step list, and (when the install included MCP servers needing
// secrets) a "set these env vars" section with copy-paste export lines.
type doneModel struct {
	palette  Palette
	bundles  []preset.Bundle
	selected map[string]bool
}

func newDone(p Palette, bundles []preset.Bundle, selected map[string]bool) doneModel {
	return doneModel{palette: p, bundles: bundles, selected: selected}
}

func (m doneModel) Init() tea.Cmd { return nil }

func (m doneModel) Update(msg tea.Msg) (doneModel, tea.Cmd) {
	if _, ok := msg.(tea.KeyMsg); ok {
		return m, tea.Quit
	}
	return m, nil
}

func (m doneModel) View(width, height int) string {
	p := m.palette
	canvasW := CanvasW(width)
	contentW := canvasW - 4

	var b strings.Builder

	// Big check + headline
	check := lipgloss.NewStyle().Foreground(p.Success).Bold(true).Render("●")
	headline := lipgloss.NewStyle().Foreground(p.Text).Bold(true).Render("WELCOME HOME")
	row := lipgloss.JoinHorizontal(lipgloss.Bottom, check, "  ", headline)
	b.WriteString(lipgloss.PlaceHorizontal(contentW, lipgloss.Center, row))
	b.WriteString("\n")
	tagline := lipgloss.NewStyle().Foreground(p.Muted).Italic(true).
		Render("your machine feels a little more like yours")
	b.WriteString(lipgloss.PlaceHorizontal(contentW, lipgloss.Center, tagline))
	b.WriteString("\n\n")

	// Next steps
	b.WriteString(SectionLabel(p, "Next steps", "", contentW))
	b.WriteString("\n\n")

	reload := reloadShellCmd()
	steps := []struct{ cmd, desc string }{
		{reload, "reload your shell (PATH updates already written to your rc)"},
		{"lfg backup", "snapshot this machine"},
		{"lfg", "re-run anytime"},
	}
	for i, s := range steps {
		num := lipgloss.NewStyle().Foreground(p.Muted).Render(strings.Repeat("0", 1) + sprint1(i+1))
		cmd := lipgloss.NewStyle().Foreground(p.Primary).Bold(true).Render(s.cmd)
		desc := lipgloss.NewStyle().Foreground(p.Muted).Render(" · " + s.desc)
		b.WriteString("  " + num + "  " + cmd + desc + "\n")
	}
	b.WriteString("\n")

	// Env vars section — listed when any installed tool declares EnvVars
	// (typically MCP servers needing GITHUB_PERSONAL_ACCESS_TOKEN etc).
	// Renders one `export FOO=...` line per var so the user can paste
	// straight into their shell rc. We never collect or write the values
	// — just remind the user what to set.
	if envs := requiredEnvVars(m.bundles, m.selected); len(envs) > 0 {
		b.WriteString(SectionLabel(p, "Env vars to set", "before MCP servers can run", contentW))
		b.WriteString("\n\n")
		boxStyle := lipgloss.NewStyle().Foreground(p.Text).Background(p.Panel).Padding(0, 1)
		for _, ev := range envs {
			b.WriteString("  " + boxStyle.Render("export "+ev+"=...") + "\n")
		}
		b.WriteString("\n")
	}

	// Star CTA + attribution. Centered so it reads as a sign-off.
	star := lipgloss.NewStyle().Foreground(p.Warn).Bold(true).Render("★")
	starMsg := lipgloss.NewStyle().Foreground(p.Text).Render("If lfg helped, star us on GitHub")
	url := lipgloss.NewStyle().Foreground(p.Accent).Underline(true).
		Render("https://github.com/ptmaroct/lfg")
	starLine := star + "  " + starMsg
	b.WriteString(lipgloss.PlaceHorizontal(contentW, lipgloss.Center, starLine))
	b.WriteString("\n")
	b.WriteString(lipgloss.PlaceHorizontal(contentW, lipgloss.Center, url))
	b.WriteString("\n\n")
	credit := lipgloss.NewStyle().Foreground(p.Muted).Italic(true).
		Render("Made by Anuj Sharma")
	b.WriteString(lipgloss.PlaceHorizontal(contentW, lipgloss.Center, credit))
	b.WriteByte('\n')

	return Frame(p, width, height,
		"all set",
		b.String(),
		HintLine(p, KeyHint(p, "any", "exit")),
		height < 22,
	)
}

// requiredEnvVars walks the install plan (bundles × selected map) and
// returns a deduped, stable-ordered list of env vars the installed
// tools declared. Used by the done screen to surface MCP secrets the
// user must set before the servers will function.
func requiredEnvVars(bundles []preset.Bundle, selected map[string]bool) []string {
	seen := map[string]bool{}
	var out []string
	for _, b := range bundles {
		for _, t := range b.Tools {
			key := b.ID + "/" + t.Name
			if !selected[key] {
				continue
			}
			for _, ev := range t.EnvVars {
				if seen[ev] {
					continue
				}
				seen[ev] = true
				out = append(out, ev)
			}
		}
	}
	return out
}

// reloadShellCmd returns the literal command the user should run to
// pick up the rc changes lfg just wrote. We can't reload the parent
// shell from a child process (POSIX doesn't expose that) so the
// command must replace the parent shell with a fresh login shell of
// the same kind.
//
// Detection priority:
//  1. $SHELL env var — set by every well-behaved shell at startup
//  2. parent process name via /proc/<ppid>/comm — covers
//     `docker run -it bash` where $SHELL stays unset
//  3. generic `exec "$SHELL" -l` so the user's own shell expands the
//     variable at runtime
//
// We deliberately drop the `-l` (login) flag for bash/zsh: login mode
// on Linux sources ~/.bash_profile (or ~/.profile / ~/.zprofile), NOT
// ~/.bashrc / ~/.zshrc — but our shellrc.go writes the PATH block to
// the interactive rc, so a login-shell exec wouldn't pick it up. Plain
// `exec bash` / `exec zsh` re-runs as interactive (since stdin is a
// tty) and sources the right file.
func reloadShellCmd() string {
	shell := strings.ToLower(filepath.Base(os.Getenv("SHELL")))
	if shell == "" || shell == "." {
		if data, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", os.Getppid())); err == nil {
			shell = strings.ToLower(strings.TrimSpace(string(data)))
		}
	}
	switch {
	case strings.Contains(shell, "fish"):
		return "exec fish"
	case strings.Contains(shell, "zsh"):
		return "exec zsh"
	case strings.Contains(shell, "bash"):
		return "exec bash"
	}
	return `exec "$SHELL"`
}
