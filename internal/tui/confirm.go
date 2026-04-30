package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/anuj/lfg/internal/preset"
)

// confirmModel — telemetry-style review screen. Big numerals as stat
// readouts, source breakdown table, then a single Y/N confirm prompt
// rendered inline (no huh form — kept under the stats so the eye lands
// on the numbers first, decision second).
type confirmModel struct {
	palette   Palette
	cursor    int // 0 = install, 1 = back
	toInstall []preset.Tool
	alreadyOK int
	bySource  map[string]int
}

func newConfirm(p Palette, bundles []preset.Bundle, selected map[string]bool) confirmModel {
	m := confirmModel{palette: p, bySource: map[string]int{}}
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
	return m
}

func (m confirmModel) Init() tea.Cmd { return nil }

func (m confirmModel) Update(msg tea.Msg) (confirmModel, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "left", "h", "right", "l", "tab":
			m.cursor = 1 - m.cursor
		case "enter":
			if m.cursor == 0 {
				return m, goTo(screenProgress)
			}
			return m, goTo(screenTree)
		case "y", "Y":
			return m, goTo(screenProgress)
		case "n", "N", "esc":
			return m, goTo(screenTree)
		}
	}
	return m, nil
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

	// Section
	b.WriteString(SectionLabel(p, "Ready to install", "review summary", contentW))
	b.WriteString("\n\n")

	// Big numerals row
	b.WriteString(renderStatRow(p, contentW, []statCell{
		{label: "TO INSTALL", value: fmt.Sprintf("%02d", len(m.toInstall)), color: p.Primary},
		{label: "ALREADY OK", value: fmt.Sprintf("%02d", m.alreadyOK), color: p.Success},
		{label: "SOURCES", value: fmt.Sprintf("%02d", len(m.bySource)), color: p.Accent},
	}))
	b.WriteString("\n\n")

	// Source breakdown table
	b.WriteString("  " + Hairline(p, contentW-2) + "\n")
	b.WriteString(renderSources(p, m.bySource, contentW))
	b.WriteString("\n")
	b.WriteString("  " + Hairline(p, contentW-2) + "\n\n")

	// Preview list (first 6)
	if len(m.toInstall) > 0 {
		previewTitle := lipgloss.NewStyle().Foreground(p.Muted).Render("  PREVIEW")
		b.WriteString(previewTitle + "\n")
		limit := 6
		if len(m.toInstall) < limit {
			limit = len(m.toInstall)
		}
		for i := 0; i < limit; i++ {
			t := m.toInstall[i]
			dot := lipgloss.NewStyle().Foreground(p.Primary).Render("  ●")
			name := lipgloss.NewStyle().Foreground(p.Text).Render(t.Name)
			src := lipgloss.NewStyle().Foreground(p.Muted).Render(t.Source)
			b.WriteString(fmt.Sprintf("%s  %s  %s\n",
				dot, padName(name, t.Name, 26), src))
		}
		if len(m.toInstall) > limit {
			more := lipgloss.NewStyle().Foreground(p.Muted).Italic(true).
				Render(fmt.Sprintf("     ... and %d more", len(m.toInstall)-limit))
			b.WriteString(more + "\n")
		}
		b.WriteString("\n")
	}

	// Choice buttons
	b.WriteString(renderConfirmButtons(p, m.cursor))
	b.WriteString("\n\n")

	// Prototype note
	warn := lipgloss.NewStyle().Foreground(p.Warn).Bold(true).Render("⚠ ")
	note := lipgloss.NewStyle().Foreground(p.Muted).Italic(true).
		Render("UX prototype — installers mocked, no system change.")
	b.WriteString("  " + warn + note)

	return Frame(p, width, height,
		"step 2/2 · confirm",
		b.String(),
		HintLine(p,
			KeyHint(p, "←→", "switch"),
			KeyHint(p, "Y", "install"),
			KeyHint(p, "N", "back"),
			KeyHint(p, "⏎", "select"),
		),
		height < 22,
	)
}

// statCell — one big numeral + label, tactical-readout style.
type statCell struct {
	label string
	value string
	color lipgloss.Color
}

func renderStatRow(p Palette, contentW int, cells []statCell) string {
	if len(cells) == 0 {
		return ""
	}
	cellW := (contentW - 2) / len(cells)
	var rendered []string
	for _, c := range cells {
		val := lipgloss.NewStyle().
			Foreground(c.color).
			Bold(true).
			Render(c.value)
		valLine := lipgloss.PlaceHorizontal(cellW, lipgloss.Center, val)
		labelLine := lipgloss.PlaceHorizontal(cellW, lipgloss.Center,
			lipgloss.NewStyle().Foreground(p.Muted).Render(c.label))
		rendered = append(rendered, lipgloss.JoinVertical(lipgloss.Left, valLine, labelLine))
	}
	return "  " + lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

func renderSources(p Palette, m map[string]int, contentW int) string {
	if len(m) == 0 {
		return lipgloss.NewStyle().Foreground(p.Muted).Italic(true).
			Render("  no sources")
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
	sep := lipgloss.NewStyle().Foreground(p.Subtle).Render("    ")
	return "  " + lipgloss.NewStyle().Foreground(p.Muted).Render("VIA  ") + strings.Join(parts, sep)
}

func renderConfirmButtons(p Palette, cursor int) string {
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
	left := make("[Y]  INSTALL", cursor == 0)
	right := make("[N]  BACK", cursor == 1)
	row := lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
	return "  " + row
}
