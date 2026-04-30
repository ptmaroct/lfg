// snap — render any TUI screen to stdout with TrueColor ANSI codes.
// Used by tools/snap-widths.sh to feed `freeze` for static PNG generation.
//
// Usage: ./snap <screen> <width> <height> [theme]
//   screens: welcome, tree, confirm, progress, done, backup
//   theme:   lfg (default), dracula, catppuccin
package main

import (
	"fmt"
	"os"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/anuj/lfg/internal/tui"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "usage: snap <screen> <width> <height> [theme]")
		os.Exit(2)
	}

	screen := os.Args[1]
	w, err := strconv.Atoi(os.Args[2])
	check(err)
	h, err := strconv.Atoi(os.Args[3])
	check(err)
	theme := tui.ThemeLFG
	if len(os.Args) > 4 {
		theme = tui.ThemeName(os.Args[4])
	}

	// Force TrueColor so freeze sees the gradient.
	lipgloss.SetColorProfile(termenv.TrueColor)

	var m tea.Model = tui.New(theme)
	_ = m.Init()
	m, _ = m.Update(tea.WindowSizeMsg{Width: w, Height: h})

	keys := keysForScreen(screen)
	for _, k := range keys {
		var msg tea.Msg
		switch k {
		case "enter":
			msg = tea.KeyMsg{Type: tea.KeyEnter}
		case "down":
			msg = tea.KeyMsg{Type: tea.KeyDown}
		case "right":
			msg = tea.KeyMsg{Type: tea.KeyRight}
		case "space":
			msg = tea.KeyMsg{Type: tea.KeySpace}
		default:
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
		}
		var cmd tea.Cmd
		m, cmd = m.Update(msg)
		if cmd != nil {
			if next := cmd(); next != nil {
				m, _ = m.Update(next)
			}
		}
	}

	fmt.Print(m.View())
}

func keysForScreen(s string) []string {
	switch s {
	case "welcome":
		return nil
	case "tree":
		return []string{"enter", "right"}
	case "confirm":
		// drive welcome → tree → expand → toggle a few → continue.
		return []string{"enter", "right", "down", "down", "space", "down", "space", "enter"}
	case "progress":
		// welcome → tree → expand → toggle → enter → confirm → install
		return []string{"enter", "right", "down", "space", "enter", "enter"}
	case "backup":
		return []string{"down", "enter"}
	default:
		return nil
	}
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
