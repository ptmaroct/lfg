package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/anuj/lfg/internal/detect"
	"github.com/anuj/lfg/internal/installer"
	"github.com/anuj/lfg/internal/preset"
	"github.com/anuj/lfg/internal/state"
	"github.com/anuj/lfg/internal/tui"
)

// runTUI is the default `lfg` action — launch the interactive bubbletea
// program. Called from rootCmd.RunE when no subcommand is given.
func runTUI() error {
	theme, fromFlag := resolveTheme()

	// Persist theme on first run / when --theme is used so subsequent
	// invocations open in the same theme without the flag.
	if fromFlag {
		s, _ := state.Load()
		s.Theme = string(theme)
		_ = state.Save(s) // soft-fail: don't block TUI startup on state-write
	}

	// Run detect concurrently across all preset tools so the picker
	// reflects real installed state. ProbeAll fans out goroutines and
	// is bounded per-tool by an internal timeout — usually completes
	// in under a second on a warm system.
	bundles := preset.All()
	results := detect.ProbeAll(bundles)
	bundles = detect.Apply(bundles, results)

	p := tea.NewProgram(
		tui.NewWithBundles(theme, bundles, tui.WithProgressRunner(installer.Run)),
		tea.WithAltScreen(),
	)
	final, err := p.Run()
	if err != nil {
		return fmt.Errorf("tui run: %w", err)
	}

	// On clean exit, save the theme that was active when the user quit
	// (handles in-TUI Ctrl+T cycling).
	if m, ok := final.(tui.Model); ok {
		s, _ := state.Load()
		s.Theme = string(m.Theme())
		_ = state.Save(s)
	}
	return nil
}
