package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Frame is the outer chrome applied to every screen.
//
// Aesthetic: tactical bulletin. Heavy ━ rules at top + bottom strips,
// hairline ─ rules within content. Content is left-padded; no centered
// rounded card. Header strip = brand mark + breadcrumb. Footer strip =
// key cells. The whole frame is then placed via lipgloss.Place so it
// sits centered in any terminal size.
//
// Inner widgets render their own internal hairline rules / tables;
// Frame only owns the outer two strips.
func Frame(p Palette, width, height int, subtitle, inner, footer string, compactTitle bool) string {
	canvasW := width - 4
	if canvasW > 100 {
		canvasW = 100
	}
	if canvasW < 56 {
		canvasW = 56
	}

	// Header strip: small brand mark + dot separator + crumb left, theme
	// breadcrumb right. No big figlet hero — that's only on welcome
	// (renderHero), called separately.
	brand := lipgloss.NewStyle().Foreground(p.Primary).Bold(true).Render("▌ lfg")
	dot := lipgloss.NewStyle().Foreground(p.Subtle).Render(" │ ")
	crumb := lipgloss.NewStyle().Foreground(p.Text).Render(strings.ToUpper(subtitle))
	leftStrip := brand + dot + crumb

	heavy := lipgloss.NewStyle().Foreground(p.Subtle).Render(strings.Repeat("━", canvasW))

	// Right-side breadcrumb empty for now; could carry "v0.1" or theme name.
	rightStrip := ""
	header := joinStrip(canvasW, leftStrip, rightStrip)

	// Footer strip: heavy rule + key cells.
	footerLine := footer

	// Pad inner content with 2-char left margin so headings line up with
	// the strip's brand mark.
	innerPadded := indent(inner, 0)

	body := strings.Join([]string{
		heavy,
		header,
		heavy,
		"",
		innerPadded,
		"",
		heavy,
		footerLine,
		heavy,
	}, "\n")

	// Pad every body line to canvasW. lipgloss.Place centers each line
	// individually using its own width — without uniform-width lines,
	// short lines (action rows, footer) get re-centered separately and
	// drift away from the heavy rules. Right-padding flattens the block
	// into a true rectangle so Place treats it as one unit.
	body = padLinesTo(body, canvasW)

	if width <= 0 || height <= 0 {
		return body
	}
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, body)
}

// padLinesTo right-pads every line in s to width visible columns.
func padLinesTo(s string, width int) string {
	lines := strings.Split(s, "\n")
	for i, ln := range lines {
		w := lipgloss.Width(ln)
		if w < width {
			lines[i] = ln + strings.Repeat(" ", width-w)
		}
	}
	return strings.Join(lines, "\n")
}

// joinStrip returns "  left....right  " padded to width.
func joinStrip(width int, left, right string) string {
	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	pad := width - leftW - rightW - 4
	if pad < 1 {
		pad = 1
	}
	return "  " + left + strings.Repeat(" ", pad) + right + "  "
}

// indent prepends `n` spaces to every line. Currently unused but kept for
// future per-screen offset needs (e.g. table indents).
func indent(s string, n int) string {
	if n <= 0 {
		return s
	}
	pad := strings.Repeat(" ", n)
	out := make([]string, 0)
	for _, ln := range strings.Split(s, "\n") {
		out = append(out, pad+ln)
	}
	return strings.Join(out, "\n")
}

// Hairline rule of given width.
func Hairline(p Palette, width int) string {
	return lipgloss.NewStyle().Foreground(p.Hairline).Render(strings.Repeat("─", width))
}

// SectionLabel renders an uppercase tracking-tight section header.
//
//	┃ BUNDLES                                              4 ITEMS
func SectionLabel(p Palette, label, suffix string, width int) string {
	bar := lipgloss.NewStyle().Foreground(p.Primary).Render("┃ ")
	main := lipgloss.NewStyle().Foreground(p.Text).Bold(true).Render(strings.ToUpper(label))
	suf := lipgloss.NewStyle().Foreground(p.Muted).Render(strings.ToUpper(suffix))
	mainW := lipgloss.Width(bar) + lipgloss.Width(main)
	sufW := lipgloss.Width(suf)
	pad := width - mainW - sufW
	if pad < 1 {
		pad = 1
	}
	return bar + main + strings.Repeat(" ", pad) + suf
}

// KeyHint renders a single `[KEY] LABEL` cell used in the footer.
// Bracket + uppercase reads as instrument-panel button.
func KeyHint(p Palette, key, label string) string {
	keyStyle := lipgloss.NewStyle().Foreground(p.Primary).Bold(true)
	bracket := lipgloss.NewStyle().Foreground(p.Subtle).Render
	labelStyle := lipgloss.NewStyle().Foreground(p.Muted)
	return bracket("[") + keyStyle.Render(key) + bracket("] ") + labelStyle.Render(strings.ToUpper(label))
}

// HintLine joins multiple hints, left-padded to align with header.
func HintLine(p Palette, hints ...string) string {
	sep := lipgloss.NewStyle().Foreground(p.Subtle).Render("  ")
	return "  " + strings.Join(hints, sep)
}
