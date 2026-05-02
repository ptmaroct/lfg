package cli

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ptmaroct/lfg/internal/installer"
	"github.com/ptmaroct/lfg/internal/state"
	"github.com/ptmaroct/lfg/internal/tui"
)

// runTUI is the default `lfg` action — launch the interactive bubbletea
// program. Called from rootCmd.RunE when no subcommand is given.
func runTUI() error {
	applyBg()

	// Debug logging: when --debug is set, redirect stdlib log + bubbletea
	// internal log into a timestamped file under the lfg config dir. Path
	// is printed on stderr both at startup and exit so it's discoverable
	// even when the alt-screen takes over the terminal.
	logPath, closeLog, err := openDebugLog()
	if err != nil {
		fmt.Fprintln(os.Stderr, "lfg: debug log setup failed:", err)
	}
	if closeLog != nil {
		defer closeLog()
		fmt.Fprintln(os.Stderr, "lfg: debug log →", logPath)
	}

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

	if logPath != "" {
		fmt.Fprintln(os.Stderr, "lfg: debug log saved →", logPath)
	}
	return nil
}

// openDebugLog wires stdlib log → ~/.config/lfg/logs/debug-<ts>.log when
// --debug is set. Returns the resolved path + a close func; when --debug
// is off both are zero-valued.
func openDebugLog() (string, func() error, error) {
	if !debugFlag {
		return "", nil, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", nil, fmt.Errorf("home dir: %w", err)
	}
	dir := filepath.Join(home, ".config", "lfg", "logs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", nil, fmt.Errorf("mkdir %s: %w", dir, err)
	}
	path := filepath.Join(dir, fmt.Sprintf("debug-%s.log", time.Now().Format("20060102-150405")))
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return "", nil, fmt.Errorf("open %s: %w", path, err)
	}
	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	log.Printf("lfg debug log opened (pid=%d)", os.Getpid())
	return path, f.Close, nil
}
