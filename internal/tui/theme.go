package tui

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// ThemeName is the user-selectable palette.
type ThemeName string

const (
	ThemeLFG        ThemeName = "lfg"
	ThemeDracula    ThemeName = "dracula"
	ThemeCatppuccin ThemeName = "catppuccin"
)

// Palette is the set of colors each screen pulls from. Fields use the
// TerminalColor interface so themes can plug in lipgloss.AdaptiveColor
// values that pick the right hex per terminal background. Light terminals
// were rendering most surfaces invisible (Text/Subtle/Muted were all
// tuned for dark bg) — adaptive colors fix that across every screen.
//
// Aesthetic guidance: 95% surfaces use Text + Muted + Subtle. Primary is
// the SHARP accent — cursor, headers, key glyphs, divider rules. Avoid
// scattering Primary into body content; it loses signal value.
// Gradient is reserved for the title hero block on welcome only.
type Palette struct {
	Bg       lipgloss.TerminalColor
	Panel    lipgloss.TerminalColor
	Primary  lipgloss.TerminalColor
	Accent   lipgloss.TerminalColor
	Success  lipgloss.TerminalColor
	Warn     lipgloss.TerminalColor
	Error    lipgloss.TerminalColor
	Muted    lipgloss.TerminalColor
	Text     lipgloss.TerminalColor
	Subtle   lipgloss.TerminalColor
	Hairline lipgloss.TerminalColor   // thin rule color (between Subtle and Muted)
	Gradient []lipgloss.TerminalColor // stops used for title hero blend only
}

// adaptive is a tiny helper so palettes read top-to-bottom without the
// `lipgloss.AdaptiveColor{...}` boilerplate on every field.
func adaptive(light, dark string) lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: light, Dark: dark}
}

// hex resolves a TerminalColor to a single hex string for places that
// need a literal (progress.WithGradient takes strings, not styles).
// Picks the dark variant on dark terminals, light on light terminals.
func hex(c lipgloss.TerminalColor) string {
	switch v := c.(type) {
	case lipgloss.AdaptiveColor:
		if lipgloss.HasDarkBackground() {
			return v.Dark
		}
		return v.Light
	case lipgloss.Color:
		return string(v)
	}
	return ""
}

// gradientColors resolves the palette's gradient stops to concrete
// lipgloss.Color values for the current terminal background. Used by
// the hand-rolled blend code in title.go which interpolates in sRGB
// space and needs literal hex strings.
func gradientColors(p Palette) []lipgloss.Color {
	out := make([]lipgloss.Color, len(p.Gradient))
	for i, c := range p.Gradient {
		out[i] = lipgloss.Color(hex(c))
	}
	return out
}

// PaletteFor returns the palette for a theme name, with lfg as default.
// Each color is an AdaptiveColor: Light side tuned for white terminals,
// Dark side tuned for dark terminals. Same theme identity, just legible
// on both.
func PaletteFor(name ThemeName) Palette {
	switch name {
	case ThemeDracula:
		return Palette{
			Bg:       adaptive("#FAFAFA", "#282A36"),
			Panel:    adaptive("#F0F0F4", "#1E1F29"),
			Primary:  adaptive("#C71585", "#FF79C6"),
			Accent:   adaptive("#6740C4", "#BD93F9"),
			Success:  adaptive("#108950", "#50FA7B"),
			Warn:     adaptive("#9A6B00", "#F1FA8C"),
			Error:    adaptive("#C0392B", "#FF5555"),
			Muted:    adaptive("#3D4A6E", "#6272A4"),
			Text:     adaptive("#1A1B26", "#F8F8F2"),
			Subtle:   adaptive("#C8CADA", "#44475A"),
			Hairline: adaptive("#D8DAEA", "#3D3F4F"),
			Gradient: []lipgloss.TerminalColor{
				adaptive("#C71585", "#FF79C6"),
				adaptive("#6740C4", "#BD93F9"),
				adaptive("#1F8FB8", "#8BE9FD"),
				adaptive("#108950", "#50FA7B"),
			},
		}
	case ThemeCatppuccin:
		return Palette{
			Bg:       adaptive("#EFF1F5", "#1E1E2E"),
			Panel:    adaptive("#E6E9EF", "#181825"),
			Primary:  adaptive("#D02A8A", "#F5C2E7"),
			Accent:   adaptive("#7C3FBF", "#CBA6F7"),
			Success:  adaptive("#1F7C3D", "#A6E3A1"),
			Warn:     adaptive("#925D00", "#F9E2AF"),
			Error:    adaptive("#C0292B", "#F38BA8"),
			Muted:    adaptive("#4C4F69", "#6C7086"),
			Text:     adaptive("#1E1E2E", "#CDD6F4"),
			Subtle:   adaptive("#BCC0CC", "#313244"),
			Hairline: adaptive("#CCD0DA", "#2A2B3D"),
			Gradient: []lipgloss.TerminalColor{
				adaptive("#D02A8A", "#F5C2E7"),
				adaptive("#7C3FBF", "#CBA6F7"),
				adaptive("#3F6BC4", "#89B4FA"),
				adaptive("#1F7C3D", "#A6E3A1"),
			},
		}
	default: // ThemeLFG
		return Palette{
			Bg:       adaptive("#FAFAFA", "#0A0A0F"),
			Panel:    adaptive("#F1F1F6", "#14141F"),
			Primary:  adaptive("#C71585", "#FF5FD9"),
			Accent:   adaptive("#5B3CD0", "#7D56F4"),
			Success:  adaptive("#00875A", "#04E5AE"),
			Warn:     adaptive("#9A6B00", "#FFBB33"),
			Error:    adaptive("#C0292B", "#FF5555"),
			Muted:    adaptive("#4A4A52", "#7A7A85"),
			Text:     adaptive("#1A1A23", "#F0F0F5"),
			Subtle:   adaptive("#C7C7D2", "#2A2A3A"),
			Hairline: adaptive("#D8D8E0", "#1F1F2E"),
			Gradient: []lipgloss.TerminalColor{
				adaptive("#C71585", "#FF5FD9"),
				adaptive("#7C3FBF", "#B37CFF"),
				adaptive("#5B3CD0", "#7D56F4"),
				adaptive("#00875A", "#04E5AE"),
			},
		}
	}
}

// HuhTheme returns a *huh.Theme configured from this palette.
// Built by cloning ThemeCharm and overriding the colors we care about —
// avoids rebuilding every style from scratch.
func HuhTheme(p Palette) *huh.Theme {
	t := huh.ThemeCharm()

	// Kill the huh group's left "active" indicator border — we already
	// frame the whole TUI in our own card, so the extra vertical bar is noise.
	t.Focused.Base = t.Focused.Base.
		Border(lipgloss.HiddenBorder()).
		BorderLeft(false).
		BorderRight(false).
		BorderTop(false).
		BorderBottom(false).
		PaddingLeft(0)
	t.Focused.Title = t.Focused.Title.Foreground(p.Primary).Bold(true)
	t.Focused.NoteTitle = t.Focused.NoteTitle.Foreground(p.Primary).Bold(true)
	t.Focused.Description = t.Focused.Description.Foreground(p.Muted).Italic(true)
	t.Focused.ErrorIndicator = t.Focused.ErrorIndicator.Foreground(p.Error)
	t.Focused.ErrorMessage = t.Focused.ErrorMessage.Foreground(p.Error)
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(p.Primary)
	t.Focused.Option = t.Focused.Option.Foreground(p.Text)
	t.Focused.MultiSelectSelector = t.Focused.MultiSelectSelector.Foreground(p.Primary)
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(p.Success)
	t.Focused.SelectedPrefix = t.Focused.SelectedPrefix.Foreground(p.Success).SetString("✓ ")
	t.Focused.UnselectedOption = t.Focused.UnselectedOption.Foreground(p.Text)
	t.Focused.UnselectedPrefix = t.Focused.UnselectedPrefix.Foreground(p.Muted).SetString("○ ")
	// Focused button: solid Primary fill with high-contrast text. Black
	// on light terminals (where Primary is darker pink) reads cleanly;
	// black on dark terminals' bright pink is also high contrast.
	t.Focused.FocusedButton = t.Focused.FocusedButton.
		Foreground(adaptive("#FFFFFF", "#000000")).
		Background(p.Primary).
		Bold(true).
		Padding(0, 2)
	// Blurred (unselected) button: NO background fill — just muted text
	// in brackets. Solid bg fills on both buttons made it hard to tell
	// which was active on light terminals; an unfilled blurred button
	// reads unmistakably as "not selected" regardless of bg color.
	t.Focused.BlurredButton = t.Focused.BlurredButton.
		Foreground(p.Muted).
		Background(p.Bg).
		Padding(0, 2)
	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.Foreground(p.Primary)
	t.Focused.TextInput.Placeholder = t.Focused.TextInput.Placeholder.Foreground(p.Muted)
	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(p.Accent)
	t.Focused.Card = t.Focused.Card.BorderForeground(p.Accent)

	// Blurred (inactive group) — also no border, same as focused.
	t.Blurred = t.Focused
	t.Blurred.Base = t.Blurred.Base.
		Border(lipgloss.HiddenBorder()).
		BorderLeft(false).
		BorderRight(false).
		BorderTop(false).
		BorderBottom(false).
		PaddingLeft(0)
	t.Blurred.Title = t.Blurred.Title.Foreground(p.Muted)
	t.Blurred.SelectSelector = t.Blurred.SelectSelector.Foreground(p.Subtle)
	t.Blurred.MultiSelectSelector = t.Blurred.MultiSelectSelector.Foreground(p.Subtle)

	// Help line at bottom
	t.Help.ShortKey = t.Help.ShortKey.Foreground(p.Accent)
	t.Help.ShortDesc = t.Help.ShortDesc.Foreground(p.Muted)
	t.Help.ShortSeparator = t.Help.ShortSeparator.Foreground(p.Subtle)

	return t
}
