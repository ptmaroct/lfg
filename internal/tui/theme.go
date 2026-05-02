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
	ThemeColorblind ThemeName = "colorblind"
)

// Palette is the set of colors each screen pulls from. Fields use the
// TerminalColor interface so themes can plug in either lipgloss.Color
// (constant) or lipgloss.NoColor (terminal default).
//
// Strategy: instead of switching every color via AdaptiveColor (which
// breaks when terminal background detection misfires — Ghostty + tmux
// + alt-screen often misreport), we pick colors that read on BOTH
// white and black backgrounds and leave Text/Bg/Panel untinted (NoColor)
// so the terminal's own default fg/bg shows through. Outcome: theme
// switches cycle accent identity without ever rendering invisible text.
//
// Aesthetic guidance: 95% surfaces use Text + Muted + Subtle. Primary is
// the SHARP accent — cursor, headers, key glyphs, divider rules. Avoid
// scattering Primary into body content; it loses signal value.
// Gradient is reserved for the title hero block on welcome only.
type Palette struct {
	Name     ThemeName              // active theme identifier — surfaced in header
	Bg       lipgloss.TerminalColor // usually NoColor — let terminal show through
	Panel    lipgloss.TerminalColor // usually NoColor
	Primary  lipgloss.TerminalColor
	Accent   lipgloss.TerminalColor
	Success  lipgloss.TerminalColor
	Warn     lipgloss.TerminalColor
	Error    lipgloss.TerminalColor
	Muted    lipgloss.TerminalColor // mid-gray, readable on both bgs
	Text     lipgloss.TerminalColor // NoColor → terminal default fg
	Subtle   lipgloss.TerminalColor // dimmer than Muted, still visible
	Hairline lipgloss.TerminalColor // thin rule color
	Gradient []lipgloss.TerminalColor
}

// hex resolves a TerminalColor to a single hex string for places that
// need a literal (progress.WithGradient takes strings, not styles).
// NoColor returns "" — caller should treat that as "skip".
func hex(c lipgloss.TerminalColor) string {
	if cc, ok := c.(lipgloss.Color); ok {
		return string(cc)
	}
	return ""
}

// gradientColors converts the palette's gradient to concrete
// lipgloss.Color values for the hand-rolled blend in title.go.
func gradientColors(p Palette) []lipgloss.Color {
	out := make([]lipgloss.Color, 0, len(p.Gradient))
	for _, c := range p.Gradient {
		if cc, ok := c.(lipgloss.Color); ok {
			out = append(out, cc)
		}
	}
	return out
}

// Mid-gray neutrals used by every theme. Picked from the readable-on-
// both-bgs band: dark enough to show on white, light enough on black.
//
//	muted   = #9CA3AF  reads as ~55% on both bgs (info text)
//	subtle  = #828894  secondary info, still visible on dark
//	hairline= #6E7280  rule lines — bumped from #5A5A66 which was
//	                   nearly invisible on near-black terminal bgs
//
// Don't go below ~#606060 — everything past that disappears on dark.
// Don't go above ~#A8A8A8 — washes out on light.
var (
	neutralMuted    = lipgloss.Color("#9CA3AF")
	neutralSubtle   = lipgloss.Color("#828894")
	neutralHairline = lipgloss.Color("#6E7280")
)

// neutralText is the body-copy foreground. We were using NoColor (let
// terminal pick) but some terminals — particularly inside docker /
// over ssh / under tmux + alt-screen — render the default fg as black
// even on a dark bg, leaving body text invisible. Pinning an explicit
// AdaptiveColor sidesteps the unreliable terminal-default lookup.
//
// Light bg → near-black for max contrast on white surfaces.
// Dark bg  → near-white but pulled slightly off-pure so it doesn't
//            scream against an OLED black; still > 12:1 contrast on #000.
var neutralText = lipgloss.AdaptiveColor{Light: "#1F2937", Dark: "#E5E7EB"}

// PaletteFor returns the palette for a theme name. Default = LFG.
//
// Each theme picks accent colors (Primary, Accent, Success, Warn, Error)
// in the bg-agnostic band: vivid enough that they pop on white AND
// black without retuning. Text/Bg/Panel stay NoColor so the terminal's
// own scheme drives those.
func PaletteFor(name ThemeName) Palette {
	p := paletteFor(name)
	if name == "" {
		p.Name = ThemeLFG
	} else {
		p.Name = name
	}
	return p
}

func paletteFor(name ThemeName) Palette {
	switch name {
	case ThemeColorblind:
		// IBM colorblind-safe scheme. Blue + orange + magenta — no
		// red/green pair. Source: davidmathlogic.com/colorblind/.
		return Palette{
			Bg: lipgloss.NoColor{}, Panel: lipgloss.NoColor{}, Text: neutralText,
			Primary:  lipgloss.Color("#3B82F6"), // blue
			Accent:   lipgloss.Color("#A78BFA"), // purple
			Success:  lipgloss.Color("#0EA5E9"), // sky (avoid green confusion)
			Warn:     lipgloss.Color("#F59E0B"), // orange
			Error:    lipgloss.Color("#EC4899"), // magenta-pink (avoid red)
			Muted:    neutralMuted,
			Subtle:   neutralSubtle,
			Hairline: neutralHairline,
			Gradient: []lipgloss.TerminalColor{
				lipgloss.Color("#3B82F6"),
				lipgloss.Color("#A78BFA"),
				lipgloss.Color("#F59E0B"),
			},
		}
	case ThemeDracula:
		// Mid-luminance Dracula. Original Dracula colors are tuned for
		// dark bg only — pulled toward darker variants so they read on
		// white too without losing the purple/pink identity.
		return Palette{
			Bg: lipgloss.NoColor{}, Panel: lipgloss.NoColor{}, Text: neutralText,
			Primary:  lipgloss.Color("#D63384"), // pink (darkened from #FF79C6)
			Accent:   lipgloss.Color("#7048E8"), // violet (darkened from #BD93F9)
			Success:  lipgloss.Color("#2EA043"), // green (darkened from #50FA7B)
			Warn:     lipgloss.Color("#D97706"), // orange (darkened from #FFB86C)
			Error:    lipgloss.Color("#DC2626"), // red (darkened from #FF5555)
			Muted:    neutralMuted,
			Subtle:   neutralSubtle,
			Hairline: neutralHairline,
			Gradient: []lipgloss.TerminalColor{
				lipgloss.Color("#D63384"),
				lipgloss.Color("#7048E8"),
				lipgloss.Color("#0EA5E9"),
				lipgloss.Color("#2EA043"),
			},
		}
	case ThemeCatppuccin:
		// Mid-luminance Catppuccin Mocha-flavored. Pastels darkened
		// enough to read on white without washing out.
		return Palette{
			Bg: lipgloss.NoColor{}, Panel: lipgloss.NoColor{}, Text: neutralText,
			Primary:  lipgloss.Color("#D02A8A"), // pink (darkened from #F5C2E7)
			Accent:   lipgloss.Color("#7C3FBF"), // mauve (darkened from #CBA6F7)
			Success:  lipgloss.Color("#1F7C3D"), // green (darkened from #A6E3A1)
			Warn:     lipgloss.Color("#C2702E"), // peach (darkened from #FAB387)
			Error:    lipgloss.Color("#C0392B"), // red (darkened from #F38BA8)
			Muted:    neutralMuted,
			Subtle:   neutralSubtle,
			Hairline: neutralHairline,
			Gradient: []lipgloss.TerminalColor{
				lipgloss.Color("#D02A8A"),
				lipgloss.Color("#7C3FBF"),
				lipgloss.Color("#3F6BC4"),
				lipgloss.Color("#1F7C3D"),
			},
		}
	default: // ThemeLFG
		// Mid-luminance vivid palette. Each accent picked to clear the
		// ~5:1 contrast bar against BOTH white and black so they don't
		// wash out on light terminals or vanish on dark ones. Hot-pink
		// brand identity preserved, but pulled darker (#E91E63) than
		// the original #FF5FD9 which faded into pastel on white bg.
		return Palette{
			Bg: lipgloss.NoColor{}, Panel: lipgloss.NoColor{}, Text: neutralText,
			Primary:  lipgloss.Color("#E91E63"), // pink-500
			Accent:   lipgloss.Color("#7C3AED"), // violet-600
			Success:  lipgloss.Color("#059669"), // emerald-600
			Warn:     lipgloss.Color("#D97706"), // amber-600
			Error:    lipgloss.Color("#DC2626"), // red-600
			Muted:    neutralMuted,
			Subtle:   neutralSubtle,
			Hairline: neutralHairline,
			Gradient: []lipgloss.TerminalColor{
				lipgloss.Color("#E91E63"),
				lipgloss.Color("#7C3AED"),
				lipgloss.Color("#059669"),
			},
		}
	}
}

// HuhTheme returns a *huh.Theme configured from this palette.
// Built by cloning ThemeCharm and overriding the colors we care about —
// avoids rebuilding every style from scratch.
func HuhTheme(p Palette) *huh.Theme {
	t := huh.ThemeCharm()

	t.Group.Base = t.Group.Base.
		Border(lipgloss.HiddenBorder()).
		BorderLeft(false).
		BorderRight(false).
		BorderTop(false).
		BorderBottom(false).
		PaddingLeft(0)

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
	// Focused button: solid Primary fill. White text reads on every
	// theme's primary (all are mid-saturation pinks/blues — white
	// passes WCAG AA on all of them).
	t.Focused.FocusedButton = t.Focused.FocusedButton.
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(p.Primary).
		Bold(true).
		Padding(0, 2)
	// Blurred (unselected) button: bracketed muted text, no bg. Reads
	// unmistakably as "not selected" because the focused button has a
	// solid fill and this one doesn't.
	t.Focused.BlurredButton = t.Focused.BlurredButton.
		Foreground(p.Muted).
		Background(lipgloss.NoColor{}).
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
	t.Help.ShortKey = t.Help.ShortKey.Foreground(p.Primary)
	t.Help.ShortDesc = t.Help.ShortDesc.Foreground(p.Muted)
	t.Help.ShortSeparator = t.Help.ShortSeparator.Foreground(p.Subtle)
	t.Help.FullKey = t.Help.FullKey.Foreground(p.Primary)
	t.Help.FullDesc = t.Help.FullDesc.Foreground(p.Muted)
	t.Help.FullSeparator = t.Help.FullSeparator.Foreground(p.Subtle)

	return t
}
