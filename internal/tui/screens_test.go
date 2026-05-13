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
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ptmaroct/lfg/internal/preset"
)

// TestMain freezes the pin-freshness clock + injects a deterministic
// pin set so welcome-screen snapshots don't drift with wall time.
// Without this every snapshot would expire daily as the embedded
// BumpedAt aged.
func TestMain(m *testing.M) {
	fixedNow, _ := time.Parse(time.RFC3339, "2026-05-13T12:00:00Z")
	restoreClock := preset.SetNowForTest(fixedNow)
	defer restoreClock()
	restorePins := preset.SetPinsForTest(preset.PinSet{
		BumpedAt: fixedNow.Add(-3 * 24 * time.Hour), // "fresh" bucket
		Pins:     map[string]preset.PinEntry{"placeholder": {Version: "x"}},
	})
	defer restorePins()
	os.Exit(m.Run())
}

// presetAllForTest returns the same hardcoded bundle data the TUI uses
// in tests (preset.All). Hoisted to its own function so multiple tests
// can share it without recreating the import dance.
func presetAllForTest() []preset.Bundle { return preset.FilterForHost(preset.All()) }

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

// TestSnapshot_QuitConfirm covers the global `q` quit dialog. Without
// this test the dialog drifted off-center for weeks because no other
// snapshot exercised the quitConfirm screen.
func TestSnapshot_QuitConfirm(t *testing.T) {
	for _, sz := range sizes {
		name := "quit_confirm_lfg_" + sz.name
		t.Run(name, func(t *testing.T) {
			driveAndSnapshot(t, name, ThemeLFG, sz, "q")
		})
	}
}

// TestSnapshot_Probe covers the first-paint detection screen. It bypasses
// the standard driveAndSnapshot helper because `New(theme)` lands on
// welcome (deterministic for the bundle-picker test) and the live
// probeModel runs a goroutine. Constructing newProbe() directly without
// calling Init() leaves the goroutine dormant, so we get a stable
// "0 of N" initial frame.
func TestSnapshot_Probe(t *testing.T) {
	for _, sz := range sizes {
		name := "probe_lfg_" + sz.name
		t.Run(name, func(t *testing.T) {
			pal := PaletteFor(ThemeLFG)
			m := newProbe(pal, presetAllForTest())
			got := m.View(sz.w, sz.h)

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
		})
	}
}

// TestSnapshot_Aliases covers the alias picker — final optional screen
// before the done card. Bypasses driveAndSnapshot because reaching the
// alias screen via real key input would require driving the whole
// install flow (slow, brittle): we construct aliasesModel directly with
// the default catalog, then render.
func TestSnapshot_Aliases(t *testing.T) {
	for _, sz := range sizes {
		name := "aliases_lfg_" + sz.name
		t.Run(name, func(t *testing.T) {
			pal := PaletteFor(ThemeLFG)
			m := newAliases(pal, preset.DefaultAliases(), nil, nil)
			pumpAliasesInit(&m)
			renderGolden(t, name, m.View(sz.w, sz.h))
		})
	}
}

// TestSnapshot_AliasesConflict locks in the rendering of the conflict
// warning suffix on alias options whose names already exist in the
// user's rc files outside the lfg-managed block.
func TestSnapshot_AliasesConflict(t *testing.T) {
	sz := termSize{"md_100x30", 100, 30}
	name := "aliases_conflict_lfg_" + sz.name
	t.Run(name, func(t *testing.T) {
		pal := PaletteFor(ThemeLFG)
		conflicts := map[string]string{
			"gd": ".zshrc:42",
			"gs": ".bashrc:17",
		}
		m := newAliases(pal, preset.DefaultAliases(), nil, conflicts)
		pumpAliasesInit(&m)
		renderGolden(t, name, m.View(sz.w, sz.h))
	})
}

// pumpAliasesInit fires the form's Init() command so huh's MultiSelect
// transitions out of its uninitialized "blank canvas" state and the
// option list actually renders for the snapshot. Without this every
// golden contains an empty form area.
func pumpAliasesInit(m *aliasesModel) {
	if cmd := m.Init(); cmd != nil {
		if msg := cmd(); msg != nil {
			updated, _ := m.Update(msg)
			*m = updated
		}
	}
}

func renderGolden(t *testing.T, name, got string) {
	t.Helper()
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
