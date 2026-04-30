package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	figure "github.com/common-nighthawk/go-figure"
)

// RenderTitle produces the big gradient banner for `lfg`.
// When `compact` is true (small terminals) it returns a single-line
// gradient-colored "lfg" instead of figlet art.
func RenderTitle(p Palette, text string, compact bool) string {
	if compact {
		return renderGradientString(p.Gradient, strings.ToUpper(text), true)
	}

	// Bundled fonts: standard, slant, big, doom, larry3d, rectangles, roman,
	// smslant, banner, basic, bell, big, block, bubble, digital, ivrit, lean,
	// mini, mnemonic, script, shadow, term. "big" reads cleanest for short
	// brand names; "slant" was too slashy for "lfg".
	fig := figure.NewFigure(text, "standard", true)
	raw := fig.String()

	// Strip trailing blank lines figure likes to pad with.
	raw = strings.TrimRight(raw, "\n ")

	// Gradient each non-space column across the width for a horizontal blend.
	lines := strings.Split(raw, "\n")
	maxW := 0
	for _, ln := range lines {
		if lipgloss.Width(ln) > maxW {
			maxW = lipgloss.Width(ln)
		}
	}
	if maxW == 0 {
		return ""
	}

	blend := lipgloss.Color("#FF5FD9")
	_ = blend
	colors := blend1D(maxW, p.Gradient...)

	var out strings.Builder
	for li, ln := range lines {
		visible := 0
		for i, r := range ln {
			if r == ' ' {
				out.WriteRune(r)
				visible++
				continue
			}
			c := colors[min2(i, len(colors)-1)]
			style := lipgloss.NewStyle().Foreground(c).Bold(true)
			out.WriteString(style.Render(string(r)))
			visible++
		}
		// Right-pad to maxW so every line has identical visible width.
		// Without this, downstream Align(Center) re-centers each line on
		// its own width, drifting the figlet shape apart.
		if pad := maxW - visible; pad > 0 {
			out.WriteString(strings.Repeat(" ", pad))
		}
		if li < len(lines)-1 {
			out.WriteByte('\n')
		}
	}
	return out.String()
}

// renderGradientString colors each non-space rune on a horizontal gradient.
func renderGradientString(stops []lipgloss.Color, s string, bold bool) string {
	if len(s) == 0 {
		return ""
	}
	colors := blend1D(len(s), stops...)
	var b strings.Builder
	for i, r := range s {
		style := lipgloss.NewStyle().Foreground(colors[i])
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
	// Split N across (len(stops)-1) segments
	segments := len(stops) - 1
	for i := 0; i < n; i++ {
		// position in [0, segments]
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
