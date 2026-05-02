package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/ptmaroct/lfg/internal/preset"
)

// confirmModel — review-and-go screen.
//
// Layout (top → bottom):
//   1. SectionLabel "READY TO INSTALL" with "STEP 2/2" suffix
//   2. Bordered stat cells: TO INSTALL · ALREADY OK · SOURCES
//   3. Source breakdown: VIA  N BREW · N CUSTOM · N MISE
//   4. Two-column preview list (collapses to one column on narrow widths)
//   5. huh.NewConfirm form (Title disabled — section label owns the
//      headline, form just renders the buttons)
type confirmModel struct {
	palette Palette
	form    *huh.Form
	// answer is "go back" (true=back, false=install). The Affirmative
	// (left) slot is "Back" so back sits on the left of the button row;
	// Install is on the right and is the default-focused button.
	// Pointer (not value) so the form's bound accessor and the model's
	// reads stay coherent across bubbletea's value-receiver copies.
	answer    *bool
	toInstall []preset.Tool
	alreadyOK int
	bySource  map[string]int
}

func newConfirm(p Palette, bundles []preset.Bundle, selected map[string]bool) confirmModel {
	// answer=false → Install (Negative slot, right) is focused so users
	// can hit Enter immediately to install.
	answer := false
	m := confirmModel{palette: p, bySource: map[string]int{}, answer: &answer}
	for _, b := range bundles {
		for _, t := range b.Tools {
			key := b.ID + "/" + t.Name
			if !selected[key] {
				if t.Installed {
					m.alreadyOK++
				}
				continue
			}
			m.toInstall = append(m.toInstall, t)
			m.bySource[t.Source]++
		}
	}

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Affirmative("Back").
				Negative("Install").
				Value(m.answer),
		),
	).
		WithTheme(HuhTheme(p)).
		WithShowHelp(false).
		WithShowErrors(false)
	return m
}

func (m confirmModel) Init() tea.Cmd { return m.form.Init() }

func (m confirmModel) Update(msg tea.Msg) (confirmModel, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "esc":
			return m, goTo(screenTree)
		case "i", "I":
			return m, goTo(screenProgress)
		case "b", "B":
			return m, goTo(screenTree)
		}
	}

	f, cmd := m.form.Update(msg)
	if ff, ok := f.(*huh.Form); ok {
		m.form = ff
	}
	if m.form.State == huh.StateCompleted {
		// answer is bound to Affirmative ("Back"). true=back, false=install.
		if m.answer != nil && *m.answer {
			return m, goTo(screenTree)
		}
		return m, goTo(screenProgress)
	}
	return m, cmd
}

func (m confirmModel) View(width, height int) string {
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

	b.WriteString(SectionLabel(p, "Ready to install", "step 2/2", contentW))
	b.WriteString("\n\n")

	// Bordered stat cells — each cell sits in its own framed box so the
	// numerals read as discrete telemetry readouts.
	b.WriteString(renderStatCells(p, contentW, []statCell{
		{label: "TO INSTALL", value: fmt.Sprintf("%02d", len(m.toInstall)), color: p.Primary},
		{label: "ALREADY OK", value: fmt.Sprintf("%02d", m.alreadyOK), color: p.Success},
		{label: "SOURCES", value: fmt.Sprintf("%02d", len(m.bySource)), color: p.Accent},
	}))
	b.WriteString("\n\n")

	// Sources line (no enclosing rules — the boxed stat cells already
	// give the section visual weight; more rules would crowd it).
	b.WriteString(renderSources(p, m.bySource, contentW))
	b.WriteString("\n\n")

	// Two-column preview list (single column when terminal narrow).
	if len(m.toInstall) > 0 {
		b.WriteString(renderPreviewColumns(p, m.toInstall, contentW))
		b.WriteString("\n")
	}

	// huh form — buttons only. Section label already supplies headline.
	b.WriteString("  " + m.form.View())

	return Frame(p, width, height,
		"step 2/2 · confirm",
		b.String(),
		HintLine(p,
			KeyHint(p, "←→", "switch"),
			KeyHint(p, "I", "install"),
			KeyHint(p, "B", "back"),
			KeyHint(p, "⏎", "select"),
		),
		height < 22,
	)
}

// statCell — one numeric readout: big colored value + uppercase muted label.
type statCell struct {
	label string
	value string
	color lipgloss.Color
}

// renderStatCells lays out cells side-by-side with rounded borders so each
// reads as its own readout. Falls back to flat numerals when columns
// would be too narrow to box cleanly.
func renderStatCells(p Palette, contentW int, cells []statCell) string {
	if len(cells) == 0 {
		return ""
	}
	gap := 2
	cellW := (contentW - 2 - gap*(len(cells)-1)) / len(cells)
	if cellW < 14 {
		return renderStatRow(p, contentW, cells)
	}

	rendered := make([]string, 0, len(cells))
	for _, c := range cells {
		val := lipgloss.NewStyle().Foreground(c.color).Bold(true).Render(c.value)
		label := lipgloss.NewStyle().Foreground(p.Muted).Render(c.label)
		body := lipgloss.JoinVertical(lipgloss.Center,
			lipgloss.PlaceHorizontal(cellW-4, lipgloss.Center, val),
			lipgloss.PlaceHorizontal(cellW-4, lipgloss.Center, label),
		)
		box := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(p.Subtle).
			Padding(0, 1).
			Width(cellW - 2).
			Render(body)
		rendered = append(rendered, box)
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, joinWithGap(rendered, gap)...)
	return "  " + row
}

func joinWithGap(items []string, gap int) []string {
	if len(items) <= 1 {
		return items
	}
	pad := strings.Repeat(" ", gap)
	out := make([]string, 0, len(items)*2-1)
	for i, it := range items {
		if i > 0 {
			out = append(out, pad)
		}
		out = append(out, it)
	}
	return out
}

// renderStatRow — flat fallback for very narrow widths.
func renderStatRow(p Palette, contentW int, cells []statCell) string {
	cellW := (contentW - 2) / len(cells)
	var rendered []string
	for _, c := range cells {
		val := lipgloss.NewStyle().Foreground(c.color).Bold(true).Render(c.value)
		valLine := lipgloss.PlaceHorizontal(cellW, lipgloss.Center, val)
		labelLine := lipgloss.PlaceHorizontal(cellW, lipgloss.Center,
			lipgloss.NewStyle().Foreground(p.Muted).Render(c.label))
		rendered = append(rendered, lipgloss.JoinVertical(lipgloss.Left, valLine, labelLine))
	}
	return "  " + lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

func renderSources(p Palette, m map[string]int, contentW int) string {
	if len(m) == 0 {
		return lipgloss.NewStyle().Foreground(p.Muted).Italic(true).Render("  no sources")
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		kStyle := lipgloss.NewStyle().Foreground(p.Text).Bold(true).Render(strings.ToUpper(k))
		vStyle := lipgloss.NewStyle().Foreground(p.Primary).Bold(true).Render(fmt.Sprintf("%d", m[k]))
		parts = append(parts, vStyle+" "+kStyle)
	}
	sep := lipgloss.NewStyle().Foreground(p.Subtle).Render(" · ")
	return "  " + lipgloss.NewStyle().Foreground(p.Muted).Render("VIA  ") + strings.Join(parts, sep)
}

// renderPreviewColumns — labelled list of tools to install. Always
// single-column so the labelled columns line up vertically. Layout:
//
//	PREVIEW (N total)              CURRENT      VIA
//	● git                          v2.45.2      brew
//	● node                         —            mise
//
// "CURRENT" = version installed on host now (— when missing).
// "VIA"     = installer backend that will handle the install.
func renderPreviewColumns(p Palette, tools []preset.Tool, contentW int) string {
	const (
		nameW    = 28
		currentW = 12
	)

	headerStyle := lipgloss.NewStyle().Foreground(p.Muted)
	header := fmt.Sprintf("  %s   %s  %s",
		headerStyle.Render(padRightPlain(fmt.Sprintf("PREVIEW  (%d total)", len(tools)), nameW)),
		headerStyle.Render(padRightPlain("CURRENT", currentW)),
		headerStyle.Render("VIA"),
	)

	limit := 8
	if len(tools) < limit {
		limit = len(tools)
	}
	more := len(tools) - limit

	row := func(t preset.Tool) string {
		dot := lipgloss.NewStyle().Foreground(p.Primary).Render("●")
		name := lipgloss.NewStyle().Foreground(p.Text).Render(t.Name)
		nameCol := padName(name, t.Name, nameW)

		var current string
		if t.Installed {
			v := t.Version
			if v == "" {
				v = "installed"
			} else {
				v = "v" + v
			}
			current = lipgloss.NewStyle().Foreground(p.Success).Render(v)
		} else {
			current = lipgloss.NewStyle().Foreground(p.Muted).Render("—")
		}
		currentCol := padPlain(current, currentW)

		via := lipgloss.NewStyle().Foreground(p.Muted).Render(t.Source)

		return fmt.Sprintf("%s %s   %s  %s", dot, nameCol, currentCol, via)
	}

	var b strings.Builder
	b.WriteString(header + "\n")
	for i := 0; i < limit; i++ {
		b.WriteString("  " + row(tools[i]) + "\n")
	}
	if more > 0 {
		b.WriteString("  " + lipgloss.NewStyle().Foreground(p.Muted).Italic(true).
			Render(fmt.Sprintf("    ... and %d more", more)))
	}
	return b.String()
}
