package tui

import (
	"strings"

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
	// answer is a heap pointer so the form's bound accessor and the
	// model's reads stay coherent across the value-receiver copies that
	// bubbletea makes on every Update. Using a plain bool here would
	// leave the form mutating the original heap-escaped struct while
	// reads happen on a stale copy — Enter on "Yes, quit" then dismisses
	// instead of quitting.
	answer *bool
}

// quitCancelMsg signals the root model to restore the previous screen.
type quitCancelMsg struct{}

func newQuitConfirm(p Palette) quitConfirmModel {
	answer := false
	m := quitConfirmModel{palette: p, answer: &answer}
	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Quit lfg?").
				Description("Leave the installer. Nothing has been written to your system yet.").
				Affirmative("Yes, quit").
				Negative("No, stay").
				Value(m.answer),
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
		case "n", "N", "esc", "backspace", "delete":
			return m, func() tea.Msg { return quitCancelMsg{} }
		}
	}

	f, cmd := m.form.Update(msg)
	if ff, ok := f.(*huh.Form); ok {
		m.form = ff
	}
	if m.form.State == huh.StateCompleted {
		if m.answer != nil && *m.answer {
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

	// Center every line of the form view independently within canvasW so
	// the title, description, and button row all sit on the canvas axis.
	// PlaceHorizontal on the whole block centers by max line width, which
	// drifts the buttons off-center because the description line is
	// wider than the button row.
	formView := m.form.View()
	lines := strings.Split(formView, "\n")
	for i, ln := range lines {
		// Strip both-side padding huh adds via its Width().Align()
		// styling; without trimming, PlaceHorizontal centers the
		// bounding box (padded) instead of the visible content, so the
		// buttons drift off the canvas axis.
		trimmed := strings.TrimSpace(ln)
		lines[i] = lipgloss.PlaceHorizontal(canvasW, lipgloss.Center, trimmed)
	}
	inner := strings.Join(lines, "\n")

	return Frame(m.palette, width, height,
		"confirm quit",
		inner,
		// No hint line — dialog buttons + Esc are self-explanatory.
		HintLine(m.palette, KeyHint(m.palette, "⎋", "dismiss")),
		height < 22,
	)
}
