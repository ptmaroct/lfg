package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// quitConfirmModel — small confirmation dialog shown when the user presses
// `q` from any screen. Y/N buttons; esc cancels and returns to the previous
// screen via a `quitCancelMsg` cmd resolved at the root.
type quitConfirmModel struct {
	palette Palette
	cursor  int // 0 = quit, 1 = cancel
}

// quitCancelMsg signals the root model to restore the previous screen.
type quitCancelMsg struct{}

func newQuitConfirm(p Palette) quitConfirmModel {
	return quitConfirmModel{palette: p}
}

func (m quitConfirmModel) Init() tea.Cmd { return nil }

func (m quitConfirmModel) Update(msg tea.Msg) (quitConfirmModel, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "left", "h", "right", "l", "tab":
			m.cursor = 1 - m.cursor
		case "y", "Y":
			return m, tea.Quit
		case "n", "N", "esc":
			return m, func() tea.Msg { return quitCancelMsg{} }
		case "enter":
			if m.cursor == 0 {
				return m, tea.Quit
			}
			return m, func() tea.Msg { return quitCancelMsg{} }
		}
	}
	return m, nil
}

func (m quitConfirmModel) View(width, height int) string {
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
	b.WriteString(SectionLabel(p, "Quit lfg?", "confirm exit", contentW))
	b.WriteString("\n\n")

	prompt := lipgloss.NewStyle().Foreground(p.Text).
		Render("Leave the installer? Any in-progress selection is preserved")
	hint := lipgloss.NewStyle().Foreground(p.Muted).Italic(true).
		Render("nothing has been written to your system yet")
	b.WriteString("  " + prompt + "\n")
	b.WriteString("  " + hint + "\n\n")

	// Buttons
	make := func(label string, active bool) string {
		base := lipgloss.NewStyle().
			Padding(0, 3).
			Border(lipgloss.NormalBorder()).
			BorderForeground(p.Subtle).
			Foreground(p.Muted)
		if active {
			base = base.
				BorderForeground(p.Primary).
				Foreground(p.Primary).
				Bold(true)
		}
		return base.Render(label)
	}
	left := make("[Y]  QUIT", m.cursor == 0)
	right := make("[N]  CANCEL", m.cursor == 1)
	row := lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
	b.WriteString("  " + row)

	return Frame(p, width, height,
		"confirm quit",
		b.String(),
		HintLine(p,
			KeyHint(p, "←→", "switch"),
			KeyHint(p, "Y", "quit"),
			KeyHint(p, "N", "cancel"),
			KeyHint(p, "⏎", "select"),
		),
		height < 22,
	)
}
