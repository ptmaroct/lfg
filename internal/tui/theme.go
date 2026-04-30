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

// Palette is the set of colors each screen pulls from.
// Every theme exposes the same keys so screens don't branch on ThemeName.
type Palette struct {
	Bg       lipgloss.Color
	Panel    lipgloss.Color
	Primary  lipgloss.Color
	Accent   lipgloss.Color
	Success  lipgloss.Color
	Warn     lipgloss.Color
	Error    lipgloss.Color
	Muted    lipgloss.Color
	Text     lipgloss.Color
	Subtle   lipgloss.Color
	Gradient []lipgloss.Color // stops used for title + border blends
}

// PaletteFor returns the palette for a theme name, with lfg as default.
func PaletteFor(name ThemeName) Palette {
	switch name {
	case ThemeDracula:
		return Palette{
			Bg:      "#282A36",
			Panel:   "#1E1F29",
			Primary: "#FF79C6", // pink
			Accent:  "#BD93F9", // purple
			Success: "#50FA7B", // green
			Warn:    "#F1FA8C",
			Error:   "#FF5555",
			Muted:   "#6272A4",
			Text:    "#F8F8F2",
			Subtle:  "#44475A",
			Gradient: []lipgloss.Color{
				"#FF79C6", "#BD93F9", "#8BE9FD", "#50FA7B",
			},
		}
	case ThemeCatppuccin:
		return Palette{
			Bg:      "#1E1E2E",
			Panel:   "#181825",
			Primary: "#F5C2E7", // pink (Mocha)
			Accent:  "#CBA6F7", // mauve
			Success: "#A6E3A1", // green
			Warn:    "#F9E2AF",
			Error:   "#F38BA8",
			Muted:   "#6C7086",
			Text:    "#CDD6F4",
			Subtle:  "#313244",
			Gradient: []lipgloss.Color{
				"#F5C2E7", "#CBA6F7", "#89B4FA", "#A6E3A1",
			},
		}
	default: // ThemeLFG
		return Palette{
			Bg:      "#0A0A0F",
			Panel:   "#14141F",
			Primary: "#FF5FD9", // hot pink
			Accent:  "#7D56F4", // charm purple
			Success: "#04E5AE", // mint
			Warn:    "#FFBB33",
			Error:   "#FF5555",
			Muted:   "#626262",
			Text:    "#F0F0F5",
			Subtle:  "#2A2A3A",
			Gradient: []lipgloss.Color{
				"#FF5FD9", "#B37CFF", "#7D56F4", "#04E5AE",
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
	t.Focused.FocusedButton = t.Focused.FocusedButton.
		Foreground(lipgloss.Color("#000000")).
		Background(p.Primary).
		Bold(true).
		Padding(0, 2)
	t.Focused.BlurredButton = t.Focused.BlurredButton.
		Foreground(p.Muted).
		Background(p.Subtle).
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
