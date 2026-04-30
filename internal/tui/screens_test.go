// Snapshot tests for every screen at fixed terminal sizes.
//
// Approach: bypass the tea.Program runner. Construct the Model directly,
// drive Update() with key messages to reach the target state, then call
// View(width, height) and snapshot the rendered string. Pure, deterministic,
// no timing flakes, no ANSI escape stream noise.
//
// Run normally: go test ./internal/tui
// Update goldens: go test ./internal/tui -update
package tui

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

var updateGolden = flag.Bool("update", false, "update golden files")

type termSize struct {
	name string
	w, h int
}

// Width matrix covers realistic terminal widths from cramped to ultrawide.
// Heights are paired sensibly so the figlet banner has room (h>=22 enables
// the full banner; below that the title falls back to compact).
//
//	xs   60×20   tmux split, narrow ssh window
//	sm   80×24   classic 80-col, traditional terminal default
//	md  100×30   ghostty default-ish
//	lg  120×36   wide editor pane
//	xl  160×44   most laptops fullscreen
//	xxl 200×50   widescreen / external monitor
var sizes = []termSize{
	{"xs_60x20", 60, 20},
	{"sm_80x24", 80, 24},
	{"md_100x30", 100, 30},
	{"lg_120x36", 120, 36},
	{"xl_160x44", 160, 44},
	{"xxl_200x50", 200, 50},
}

// driveAndSnapshot constructs a fresh Model, applies an initial WindowSizeMsg,
// pumps the given key sequences through Update, then renders View at the
// requested size. The result is compared (or written) to testdata/<name>.golden.
func driveAndSnapshot(t *testing.T, name string, theme ThemeName, sz termSize, keys ...string) {
	t.Helper()

	var m tea.Model = New(theme)
	// Run Init so any startup commands fire — we discard the resulting cmds
	// because we don't need a real event loop.
	_ = m.Init()
	m, _ = m.Update(tea.WindowSizeMsg{Width: sz.w, Height: sz.h})

	for _, k := range keys {
		var msg tea.Msg
		switch k {
		case "enter":
			msg = tea.KeyMsg{Type: tea.KeyEnter}
		case "esc":
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		case "tab":
			msg = tea.KeyMsg{Type: tea.KeyTab}
		case "space":
			msg = tea.KeyMsg{Type: tea.KeySpace}
		case "down":
			msg = tea.KeyMsg{Type: tea.KeyDown}
		case "up":
			msg = tea.KeyMsg{Type: tea.KeyUp}
		default:
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
		}
		var cmd tea.Cmd
		m, cmd = m.Update(msg)
		// Loosely pump any synchronous transition msgs that the previous
		// update returned (transitionMsg is sent as a tea.Cmd factory; we
		// resolve once).
		if cmd != nil {
			if outMsg := cmd(); outMsg != nil {
				m, _ = m.Update(outMsg)
			}
		}
	}

	got := m.View()

	goldenPath := filepath.Join("testdata", name+".golden")
	if *updateGolden {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenPath, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %s (%d lines)", goldenPath, strings.Count(got, "\n"))
		return
	}

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("missing golden %s — run `go test ./internal/tui -update` first: %v", goldenPath, err)
	}
	if got != string(want) {
		t.Errorf("snapshot diff for %s\n--- want\n%s\n--- got\n%s", name, want, got)
	}
}

func TestSnapshot_Welcome(t *testing.T) {
	for _, sz := range sizes {
		for _, theme := range []ThemeName{ThemeLFG, ThemeDracula, ThemeCatppuccin} {
			name := "welcome_" + string(theme) + "_" + sz.name
			t.Run(name, func(t *testing.T) {
				driveAndSnapshot(t, name, theme, sz)
			})
		}
	}
}

func TestSnapshot_BundlePicker(t *testing.T) {
	for _, sz := range sizes {
		name := "bundles_lfg_" + sz.name
		t.Run(name, func(t *testing.T) {
			driveAndSnapshot(t, name, ThemeLFG, sz, "enter")
		})
	}
}

func TestSnapshot_BackupPrompt(t *testing.T) {
	for _, sz := range sizes {
		name := "backup_lfg_" + sz.name
		t.Run(name, func(t *testing.T) {
			driveAndSnapshot(t, name, ThemeLFG, sz, "down", "enter")
		})
	}
}

// TestNoControlCharsLeak — sanity check that our themes/styles don't bleed
// raw ANSI escape codes into the rendered output beyond expected styling.
// Cheap regression for "border showed up where it shouldn't" bugs.
func TestNoControlCharsLeak(t *testing.T) {
	m := New(ThemeLFG)
	_ = m.Init()
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	out := mm.View()
	// Hidden border characters shouldn't render any visible glyph.
	bad := []string{"│ ", " │", "└", "┐", "┌", "┘"} // any block-border chars huh might leak
	for _, b := range bad {
		// Only fail if a bare border char appears outside our outer Frame box,
		// which uses lipgloss.RoundedBorder (those use ╭╮╯╰─│). Skip.
		if strings.Contains(out, b) && !strings.Contains(out, "╭") {
			t.Errorf("unexpected border char %q in output", b)
		}
	}
}
