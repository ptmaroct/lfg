package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// doneModel — final card.
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
	big := lipgloss.NewStyle().Foreground(p.Success).Bold(true).Render("✓")
	headline := lipgloss.NewStyle().Foreground(p.Text).Bold(true).Render("Welcome home.")
	sub := lipgloss.NewStyle().Foreground(p.Muted).Italic(true).Render("Your machine feels a little more like yours.")

	step := func(cmd, label string) string {
		c := lipgloss.NewStyle().Foreground(p.Primary).Bold(true).Render(cmd)
		l := lipgloss.NewStyle().Foreground(p.Muted).Render(" " + label)
		return c + l
	}

	nextBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.Subtle).
		Padding(1, 2).
		Width(48).
		Render(lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Foreground(p.Text).Bold(true).Render("next steps"),
			"",
			step("exec zsh", "reload your shell"),
			step("lfg backup", "snapshot this machine to a file"),
			step("lfg", "re-run anytime"),
		))

	inner := lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.NewStyle().Foreground(p.Success).Bold(true).Padding(0, 2).Render(big+" "+headline),
		sub,
		"",
		nextBox,
	)

	return Frame(p, width, height,
		"all set",
		inner,
		HintLine(p, KeyHint(p, "any key", "exit")),
		height < 22,
	)
}
