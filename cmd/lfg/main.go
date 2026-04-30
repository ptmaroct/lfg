// Command lfg is the TUI bootstrap CLI entrypoint.
//
// UX-prototype build: installers are mocked, no filesystem changes.
// Runs through every screen so the flow can be tuned.
package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/anuj/lfg/internal/tui"
)

func main() {
	themeFlag := flag.String("theme", "lfg", "color theme: lfg, dracula, catppuccin")
	flag.Parse()

	theme := tui.ThemeName(*themeFlag)
	switch theme {
	case tui.ThemeLFG, tui.ThemeDracula, tui.ThemeCatppuccin:
	default:
		fmt.Fprintf(os.Stderr, "unknown theme %q. use one of: lfg, dracula, catppuccin\n", *themeFlag)
		os.Exit(2)
	}

	p := tea.NewProgram(tui.New(theme), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "lfg:", err)
		os.Exit(1)
	}
}
