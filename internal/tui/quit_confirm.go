package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// quitConfirmModel — huh.NewConfirm dialog shown when the user presses
// `q` from any screen. Uses the proper Charm component (matches the
// rest of the form-driven flow) instead of a hand-rolled button row.
//
// Result handling:
//   - form completed + answer=true  → tea.Quit
//   - form completed + answer=false → quitCancelMsg → root restores
//     the previous screen.
type quitConfirmModel struct {
	palette Palette
	form    *huh.Form
	answer  bool
}

// quitCancelMsg signals the root model to restore the previous screen.
type quitCancelMsg struct{}

func newQuitConfirm(p Palette) quitConfirmModel {
	m := quitConfirmModel{palette: p}
	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Quit lfg?").
				Description("Leave the installer. Nothing has been written to your system yet.").
				Affirmative("Yes, quit").
				Negative("No, stay").
				Value(&m.answer),
		),
	).
		WithTheme(HuhTheme(p)).
		WithShowHelp(false).
		WithShowErrors(false)
	return m
}

func (m quitConfirmModel) Init() tea.Cmd { return m.form.Init() }

func (m quitConfirmModel) Update(msg tea.Msg) (quitConfirmModel, tea.Cmd) {
	// Quick-exit hotkeys before the form sees the message.
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "y", "Y":
			return m, tea.Quit
		case "esc":
			return m, func() tea.Msg { return quitCancelMsg{} }
		}
	}

	f, cmd := m.form.Update(msg)
	if ff, ok := f.(*huh.Form); ok {
		m.form = ff
	}
	if m.form.State == huh.StateCompleted {
		if m.answer {
			return m, tea.Quit
		}
		return m, func() tea.Msg { return quitCancelMsg{} }
	}
	return m, cmd
}

func (m quitConfirmModel) View(width, height int) string {
	// Compute the same canvas width Frame uses so the form output sits
	// inside the frame's bounding box. Without this wrap the raw
	// huh.Form view can be wider than canvasW, which makes
	// lipgloss.Place center on the form's width (not the frame's) and
	// the whole dialog drifts left in wide terminals.
	canvasW := width - 4
	if canvasW > 100 {
		canvasW = 100
	}
	if canvasW < 56 {
		canvasW = 56
	}
	contentW := canvasW - 4

	inner := lipgloss.PlaceHorizontal(contentW, lipgloss.Center, m.form.View())

	return Frame(m.palette, width, height,
		"confirm quit",
		inner,
		HintLine(m.palette,
			KeyHint(m.palette, "←→", "switch"),
			KeyHint(m.palette, "Y", "quit"),
			KeyHint(m.palette, "⎋", "cancel"),
			KeyHint(m.palette, "⏎", "select"),
		),
		height < 22,
	)
}
