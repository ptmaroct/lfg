// Package cli wires the cobra subcommand tree.
//
// Default behavior (no subcommand) = run the TUI. Subcommands like
// `lfg backup`, `lfg doctor` bypass the TUI for headless / scripted use.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/anuj/lfg/internal/state"
	"github.com/anuj/lfg/internal/tui"
)

// themeFlag holds the persistent --theme value parsed at the root.
// Stored at package level so subcommands can read it without re-parsing.
var themeFlag string

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
	case tui.ThemeLFG, tui.ThemeDracula, tui.ThemeCatppuccin:
		return t
	}
	fmt.Fprintf(os.Stderr, "lfg: unknown theme %q, falling back to 'lfg'\n", name)
	return tui.ThemeLFG
}
