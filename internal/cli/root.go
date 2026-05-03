// Package cli wires the cobra subcommand tree.
//
// Default behavior (no subcommand) = run the TUI. Subcommands like
// `lfg backup`, `lfg doctor` bypass the TUI for headless / scripted use.
package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"

	"github.com/ptmaroct/lfg/internal/preset"
	"github.com/ptmaroct/lfg/internal/state"
	"github.com/ptmaroct/lfg/internal/tui"
)

// themeFlag holds the persistent --theme value parsed at the root.
// Stored at package level so subcommands can read it without re-parsing.
var themeFlag string

// dryRun is the persistent --dry-run / -n flag. When set every
// destructive action falls back to its preview path:
//   - default TUI: installer goroutine swaps to mockProgressRunner
//   - lfg apply: prints planned commands, doesn't exec
//   - lfg backup: prints source list + would-be filename, writes nothing
//   - lfg update: prints target asset URL, doesn't swap binary (TODO)
var dryRun bool

// configFlag points at a TOML preset file (local path or http(s) URL).
// Empty → use the built-in preset.All().
var configFlag string

// bgFlag forces the terminal background interpretation: auto / dark / light.
// "auto" calls into termenv; if that fails (some terminals don't reply to
// the OSC-11 query), we fall back to dark — dev terminals skew dark.
var bgFlag string

// debugFlag enables verbose logging to ~/.config/lfg/logs/debug-<ts>.log.
// Path is printed on stderr at startup and on exit so it's easy to pipe
// into a viewer while iterating in a docker container.
var debugFlag bool

// rootCmd is the top-level `lfg` command.
var rootCmd = &cobra.Command{
	Use:   "lfg",
	Short: "Bootstrap a fresh dev machine with one command",
	Long: `lfg sets up a new dev machine — installs CLIs, restores configs,
backs up your environment to a single file. Run with no args to launch
the interactive TUI, or use subcommands for headless workflows.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	// No subcommand → run TUI.
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTUI()
	},
}

// Execute is the single entrypoint called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "lfg:", err)
		os.Exit(1)
	}
}

func init() {
	// Persistent so subcommands inherit it (e.g. `lfg apply --theme=dracula`).
	rootCmd.PersistentFlags().StringVar(&themeFlag, "theme", "",
		"color theme: lfg, dracula, catppuccin (default: persisted or 'lfg')")
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "n", false,
		"preview only — no commands run, no files written")
	rootCmd.PersistentFlags().StringVarP(&configFlag, "config", "c", "",
		"preset file (local path or http(s) URL); defaults to the built-in preset")
	rootCmd.PersistentFlags().StringVar(&bgFlag, "bg", "auto",
		"terminal background: auto, dark, light (or set LFG_BG)")
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false,
		"verbose logging to ~/.config/lfg/logs/debug-<ts>.log")
}

// applyBg resolves --bg / $LFG_BG / auto-detect into a single
// HasDarkBackground decision and pins lipgloss to it. Auto-detect uses
// termenv (which queries OSC-11). Some terminals don't reply — we
// default to dark in that case because dev terminals skew dark.
func applyBg() {
	choice := strings.ToLower(strings.TrimSpace(bgFlag))
	if env := strings.ToLower(strings.TrimSpace(os.Getenv("LFG_BG"))); env != "" {
		choice = env
	}
	switch choice {
	case "dark":
		lipgloss.SetHasDarkBackground(true)
		return
	case "light":
		lipgloss.SetHasDarkBackground(false)
		return
	}
	// auto: ask termenv. If it returns false but the terminal is dark
	// (common — many terminals don't answer OSC-11), prefer dark.
	out := termenv.NewOutput(os.Stdout)
	dark := out.HasDarkBackground()
	if !dark {
		// Some terminals (alacritty / ghostty in some configs) report
		// false even when bg is dark. As a tie-break, sniff COLORFGBG
		// when set — second field is bg ANSI; "0" is black, "8"+ are
		// dark. If absent, keep dark as the safe default.
		if cfb := os.Getenv("COLORFGBG"); cfb != "" {
			parts := strings.Split(cfb, ";")
			if len(parts) >= 2 {
				bg := parts[len(parts)-1]
				// 0–6, 8 are dark in the typical mapping; 7, 15 are light.
				dark = bg != "7" && bg != "15"
			}
		} else {
			dark = true
		}
	}
	lipgloss.SetHasDarkBackground(dark)
}

// loadPreset returns the bundle slice for this run. Honors --config when
// set, otherwise falls back to the built-in preset.All(). Errors here
// are fatal — the user explicitly asked to load a config and we don't
// want to silently fall back to defaults under their nose.
func loadPreset() ([]preset.Bundle, error) {
	if configFlag == "" {
		return preset.FilterForHost(preset.All()), nil
	}
	bundles, err := preset.Load(configFlag)
	if err != nil {
		return nil, fmt.Errorf("load preset %q: %w", configFlag, err)
	}
	return preset.FilterForHost(bundles), nil
}

// resolveTheme picks the theme to use this run. Priority:
//  1. --theme flag (explicit override)
//  2. saved state.Theme
//  3. ThemeLFG (default)
//
// Returns the resolved name plus a bool indicating whether it came from
// flag (vs persisted/default) so callers can decide to re-save.
func resolveTheme() (tui.ThemeName, bool) {
	if themeFlag != "" {
		return validateTheme(themeFlag), true
	}
	s, err := state.Load()
	if err == nil && s.Theme != "" {
		return validateTheme(s.Theme), false
	}
	return tui.ThemeLFG, false
}

// validateTheme returns the requested theme if known, else falls back to
// ThemeLFG with a stderr warning. Doesn't exit — bad theme name is a soft
// error, not worth aborting the whole bootstrap over.
func validateTheme(name string) tui.ThemeName {
	t := tui.ThemeName(name)
	switch t {
	case tui.ThemeLFG, tui.ThemeDracula, tui.ThemeCatppuccin, tui.ThemeColorblind:
		return t
	}
	fmt.Fprintf(os.Stderr, "lfg: unknown theme %q, falling back to 'lfg'\n", name)
	return tui.ThemeLFG
}
