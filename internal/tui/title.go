package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Hand-drawn ANSI Shadow logo for "lfg".
//
// Why hand-drawn instead of go-figure: the bundled figlet fonts in
// common-nighthawk/go-figure render `l f g` with disconnected verticals
// (the brand reads as random pipes). A 6-row block-shaded glyph set
// gives a clean, instantly-readable mark and lets us animate a gradient
// sweep across columns without per-rune positioning quirks.
//
// Each row is the same visible width so column-indexed coloring lines up.
var lfgLogo = []string{
	`██╗     ███████╗ ██████╗ `,
	`██║     ██╔════╝██╔════╝ `,
	`██║     █████╗  ██║  ███╗`,
	`██║     ██╔══╝  ██║   ██║`,
	`███████╗██║     ╚██████╔╝`,
	`╚══════╝╚═╝      ╚═════╝ `,
}

// Compact single-line gradient text for tiny terminals (height < 22).
func renderCompactBrand(p Palette, text string) string {
	return renderGradientString(p.Gradient, strings.ToUpper(text), 0, true)
}

// RenderTitle returns the brand mark, optionally with an animation phase
// for the gradient sweep. `phase` shifts the gradient stops horizontally
// — pass 0 for deterministic (snapshot tests), increment per tick for
// the running TUI.
func RenderTitle(p Palette, text string, compact bool, phase int) string {
	if compact {
		return renderCompactBrand(p, text)
	}

	// We render `lfg` always — `text` arg kept for API compatibility but
	// the hand-drawn glyphs only cover those three letters.
	maxW := 0
	for _, ln := range lfgLogo {
		if w := lipgloss.Width(ln); w > maxW {
			maxW = w
		}
	}
	if maxW == 0 {
		return ""
	}

	// Build a wider gradient and shift by phase so colors slide across.
	// 2× width gives smoother motion without wraparound seams.
	colors := blend1D(maxW*2, repeatStops(p.Gradient, 2)...)

	var out strings.Builder
	for li, ln := range lfgLogo {
		col := 0
		for _, r := range ln {
			if r == ' ' {
				out.WriteRune(r)
				col++
				continue
			}
			c := colors[(col+phase)%len(colors)]
			style := lipgloss.NewStyle().Foreground(c).Bold(true)
			out.WriteString(style.Render(string(r)))
			col++
		}
		// Right-pad each row so downstream center-placement treats the
		// whole block as one rectangle (otherwise short rows recenter).
		if pad := maxW - col; pad > 0 {
			out.WriteString(strings.Repeat(" ", pad))
		}
		if li < len(lfgLogo)-1 {
			out.WriteByte('\n')
		}
	}
	return out.String()
}

// repeatStops loops the palette stops `times` times so blend1D produces a
// wider strip we can offset for animation.
func repeatStops(stops []lipgloss.Color, times int) []lipgloss.Color {
	if times <= 1 {
		return stops
	}
	out := make([]lipgloss.Color, 0, len(stops)*times)
	for i := 0; i < times; i++ {
		out = append(out, stops...)
	}
	return out
}

// renderGradientString colors each non-space rune on a horizontal gradient.
// `phase` shifts the gradient for animation.
func renderGradientString(stops []lipgloss.Color, s string, phase int, bold bool) string {
	if len(s) == 0 {
		return ""
	}
	colors := blend1D(len(s)*2, repeatStops(stops, 2)...)
	var b strings.Builder
	for i, r := range s {
		c := colors[(i+phase)%len(colors)]
		style := lipgloss.NewStyle().Foreground(c)
		if bold {
			style = style.Bold(true)
		}
		b.WriteString(style.Render(string(r)))
	}
	return b.String()
}

// blend1D interpolates N colors across the provided stops.
// Simple linear interpolation in sRGB — good enough for TUI decorative use.
// lipgloss v2 has Blend1D built-in; v1 (what we have) doesn't, so we
// hand-roll a small version here.
func blend1D(n int, stops ...lipgloss.Color) []lipgloss.Color {
	if n <= 0 {
		return nil
	}
	if len(stops) == 0 {
		return nil
	}
	if len(stops) == 1 || n == 1 {
		out := make([]lipgloss.Color, n)
		for i := range out {
			out[i] = stops[0]
		}
		return out
	}

	out := make([]lipgloss.Color, n)
	segments := len(stops) - 1
	for i := 0; i < n; i++ {
		pos := float64(i) / float64(n-1) * float64(segments)
		segIdx := int(pos)
		if segIdx >= segments {
			segIdx = segments - 1
		}
		local := pos - float64(segIdx)
		a := hexToRGB(string(stops[segIdx]))
		b := hexToRGB(string(stops[segIdx+1]))
		r := uint8(float64(a[0]) + (float64(b[0])-float64(a[0]))*local)
		g := uint8(float64(a[1]) + (float64(b[1])-float64(a[1]))*local)
		bl := uint8(float64(a[2]) + (float64(b[2])-float64(a[2]))*local)
		out[i] = rgbToHex(r, g, bl)
	}
	return out
}

func hexToRGB(h string) [3]uint8 {
	h = strings.TrimPrefix(h, "#")
	if len(h) != 6 {
		return [3]uint8{255, 255, 255}
	}
	var r, g, b uint8
	for i, c := range h {
		v := hexDigit(byte(c))
		switch i {
		case 0:
			r = v << 4
		case 1:
			r |= v
		case 2:
			g = v << 4
		case 3:
			g |= v
		case 4:
			b = v << 4
		case 5:
			b |= v
		}
	}
	return [3]uint8{r, g, b}
}

func hexDigit(c byte) uint8 {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}

func rgbToHex(r, g, b uint8) lipgloss.Color {
	const hex = "0123456789ABCDEF"
	out := []byte{'#', 0, 0, 0, 0, 0, 0}
	out[1] = hex[r>>4]
	out[2] = hex[r&0xf]
	out[3] = hex[g>>4]
	out[4] = hex[g&0xf]
	out[5] = hex[b>>4]
	out[6] = hex[b&0xf]
	return lipgloss.Color(string(out))
}

func min2(a, b int) int {
	if a < b {
		return a
	}
	return b
}
