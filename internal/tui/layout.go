package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Frame is the outer chrome applied to every screen.
// Returns a fully rendered string sized to (width × height), centered.
//
//	 ╭────── ✨ lfg ✨ ──────╮
//	 │                        │
//	 │      [title art]       │
//	 │     subtitle line      │
//	 │                        │
//	 │     inner content      │
//	 │                        │
//	 │    footer key hints    │
//	 ╰────────────────────────╯
//
// Width of the inner panel is min(term_width-4, 90). Height is term_height.
func Frame(p Palette, width, height int, subtitle, inner, footer string, compactTitle bool) string {
	canvasW := width - 4
	if canvasW > 90 {
		canvasW = 90
	}
	if canvasW < 40 {
		canvasW = 40
	}

	title := RenderTitle(p, "lfg", compactTitle)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(p.Muted).
		Italic(true).
		Align(lipgloss.Center).
		Width(canvasW - 4)

	footerStyle := lipgloss.NewStyle().
		Foreground(p.Muted).
		Align(lipgloss.Center).
		Width(canvasW - 4).
		MarginTop(1)

	// PlaceHorizontal treats input as one atomic block — preserves figlet
	// art alignment. Per-line Align(Center) would drift uneven-width lines.
	titleBlock := lipgloss.PlaceHorizontal(canvasW-4, lipgloss.Center, title)

	// Wrap inner widget (huh form, summary panels, etc.) in a card so its
	// rectangle has clear edges and can be centered as one unit. Without the
	// card the form's lines were flush-left within their centered bounding
	// box, looking off-center even though they technically were centered.
	cardW := canvasW - 12
	if cardW < 40 {
		cardW = 40
	}
	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.Subtle).
		Padding(1, 3).
		Width(cardW).
		Render(strings.TrimRight(inner, "\n"))
	innerBlock := lipgloss.PlaceHorizontal(canvasW-4, lipgloss.Center, card)

	sections := []string{titleBlock}
	if subtitle != "" {
		sections = append(sections, subtitleStyle.Render(subtitle))
	}
	sections = append(sections, "")
	sections = append(sections, innerBlock)
	if footer != "" {
		sections = append(sections, footerStyle.Render(footer))
	}
	body := lipgloss.JoinVertical(lipgloss.Center, sections...)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.Accent).
		Padding(1, 2).
		Width(canvasW).
		Render(body)

	if width <= 0 || height <= 0 {
		return box
	}
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

// KeyHint renders a single `key · label` pair used in the footer.
func KeyHint(p Palette, key, label string) string {
	k := lipgloss.NewStyle().Foreground(p.Primary).Bold(true).Render(key)
	d := lipgloss.NewStyle().Foreground(p.Muted).Render(" " + label)
	return k + d
}

// HintLine joins multiple hints with a middle-dot separator.
func HintLine(p Palette, hints ...string) string {
	sep := lipgloss.NewStyle().Foreground(p.Subtle).Render(" · ")
	return strings.Join(hints, sep)
}
