package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Frame is the outer chrome applied to every screen.
//
// Aesthetic: tactical bulletin with anchored corners. Heavy corner glyphs
// (┏ ┓ ┗ ┛) bracket the top and bottom rules so the chrome reads as a
// container without verticals (verticals on every line break centering).
// Header strip = brand mark + breadcrumb. Footer strip = key cells.
// Hairlines (─) divide sections inside the body. The whole frame is then
// placed via lipgloss.Place so it sits centered in any terminal size.
//
// Inner widgets render their own internal hairline rules / tables;
// Frame only owns the outer two strips.
func Frame(p Palette, width, height int, subtitle, inner, footer string, compactTitle bool) string {
	canvasW := CanvasW(width)

	// Header strip: small brand mark + dot separator + crumb left, theme
	// breadcrumb right. No big figlet hero — that's only on welcome
	// (renderHero), called separately.
	brand := lipgloss.NewStyle().Foreground(p.Primary).Bold(true).Render("▌ lfg")
	dot := lipgloss.NewStyle().Foreground(p.Subtle).Render(" │ ")
	crumb := lipgloss.NewStyle().Foreground(p.Text).Render(strings.ToUpper(subtitle))
	leftStrip := brand + dot + crumb

	// Right-side breadcrumb: active theme name. Surfaces what `^T` cycles
	// landed on so users see the change reflected in chrome.
	themeStyle := lipgloss.NewStyle().Foreground(p.Accent).Bold(true)
	themeLabel := lipgloss.NewStyle().Foreground(p.Subtle).Render("THEME ")
	rightStrip := themeLabel + themeStyle.Render(strings.ToUpper(string(p.Name)))
	if p.Name == "" {
		rightStrip = ""
	}
	header := joinStrip(canvasW, leftStrip, rightStrip)

	// Top + bottom anchored rules with corner glyphs.
	topRule := renderCornerRule(p, canvasW, "top")
	botRule := renderCornerRule(p, canvasW, "bottom")
	hairline := lipgloss.NewStyle().Foreground(p.Hairline).Render(strings.Repeat("─", canvasW))

	footerLine := footer

	// Inner content padding (currently no extra indent — body manages its own).
	innerPadded := indent(inner, 0)

	body := strings.Join([]string{
		topRule,
		header,
		hairline,
		"",
		innerPadded,
		"",
		hairline,
		footerLine,
		botRule,
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

// renderCornerRule builds the top or bottom rule with anchored corners.
// Top:    ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
// Bottom: ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
// Corners take Primary; the bar takes Subtle for a calmer rule.
func renderCornerRule(p Palette, width int, pos string) string {
	if width < 2 {
		return ""
	}
	var lc, rc string
	switch pos {
	case "top":
		lc, rc = "┏", "┓"
	case "bottom":
		lc, rc = "┗", "┛"
	default:
		lc, rc = "━", "━"
	}
	corner := lipgloss.NewStyle().Foreground(p.Primary).Bold(true)
	bar := lipgloss.NewStyle().Foreground(p.Subtle)
	return corner.Render(lc) + bar.Render(strings.Repeat("━", width-2)) + corner.Render(rc)
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

// CanvasW is the SINGLE source of truth for inner-canvas width across
// every screen. Was duplicated in 9 places — even with identical
// formulas, having Frame and each screen compute independently meant
// any future tweak (or a one-character typo) would silently desync
// widths between screens. Centralizing here means screens visually
// align across transitions, no surprise resize when stepping welcome
// → tree → confirm → progress → done.
//
// Bounds: 56 min (anything below makes the welcome hero unreadable),
// 100 max (wider than that bloats horizontal eye travel for a
// terminal app and content has been tuned at 100 cols).
func CanvasW(width int) int {
	w := width - 4
	if w > 100 {
		w = 100
	}
	if w < 56 {
		w = 56
	}
	return w
}
