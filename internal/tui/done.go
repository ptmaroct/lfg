package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// doneModel — final celebration card. Big checkmark, headline, next-step list.
type doneModel struct {
	palette Palette
}

func newDone(p Palette) doneModel { return doneModel{palette: p} }

func (m doneModel) Init() tea.Cmd { return nil }

func (m doneModel) Update(msg tea.Msg) (doneModel, tea.Cmd) {
	if _, ok := msg.(tea.KeyMsg); ok {
		return m, tea.Quit
	}
	return m, nil
}

func (m doneModel) View(width, height int) string {
	p := m.palette
	canvasW := width - 4
	if canvasW > 100 {
		canvasW = 100
	}
	if canvasW < 56 {
		canvasW = 56
	}
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

	steps := []struct{ cmd, desc string }{
		{"exec zsh", "reload your shell"},
		{"lfg backup", "snapshot this machine"},
		{"lfg", "re-run anytime"},
	}
	for i, s := range steps {
		num := lipgloss.NewStyle().Foreground(p.Muted).Render(strings.Repeat("0", 1) + sprint1(i+1))
		cmd := lipgloss.NewStyle().Foreground(p.Primary).Bold(true).Render(s.cmd)
		desc := lipgloss.NewStyle().Foreground(p.Muted).Render(" · " + s.desc)
		b.WriteString("  " + num + "  " + cmd + desc + "\n")
	}

	return Frame(p, width, height,
		"all set",
		b.String(),
		HintLine(p, KeyHint(p, "any", "exit")),
		height < 22,
	)
}
