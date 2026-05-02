package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ptmaroct/lfg/internal/installer"
	"github.com/ptmaroct/lfg/internal/state"
	"github.com/ptmaroct/lfg/internal/tui"
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

	// Detection runs inside the TUI on the first screen (screenProbe) so
	// the user sees animated progress instead of a frozen terminal while
	// goroutines fan out. Probe finishes → transition to welcome with
	// detect-applied bundles + the harness list set on the installer pkg.
	bundles, err := loadPreset()
	if err != nil {
		return err
	}

	// In dry-run mode the TUI still walks every screen — the only
	// difference is the install step uses the mock runner (canned lines,
	// short sleeps) so no commands hit the host. Same UX, zero side
	// effects. Useful for demos + first-time poking.
	opts := []tui.Option{tui.WithProgressRunner(installer.Run)}
	if dryRun {
		opts = []tui.Option{} // empty → mockProgressRunner default
		fmt.Fprintln(os.Stderr, "lfg: dry-run mode — no commands will be executed")
	}

	p := tea.NewProgram(
		tui.NewWithProbe(theme, bundles, opts...),
		tea.WithAltScreen(),
	)
	final, runErr := p.Run()
	if runErr != nil {
		return fmt.Errorf("tui run: %w", runErr)
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
