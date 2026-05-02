package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ptmaroct/lfg/internal/preset"
)

// exportModel — "Save current bundles as TOML" dialog reachable from
// the welcome screen. Pre-fills with a sensible filename in the user's
// home directory; on submit calls preset.Save and shows a confirmation
// with the full path.
type exportModel struct {
	palette  Palette
	bundles  []preset.Bundle
	input    textinput.Model
	saved    bool
	savedAt  string
	err      string
}

func newExport(p Palette, bundles []preset.Bundle) exportModel {
	ti := textinput.New()
	ti.Prompt = "› "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(p.Primary).Bold(true)
	ti.TextStyle = lipgloss.NewStyle().Foreground(p.Text)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(p.Muted).Italic(true)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(p.Primary)
	ti.CharLimit = 1024
	ti.Width = 80
	// Default filename: ~/lfg-preset-YYYY-MM-DD.toml
	if home, err := os.UserHomeDir(); err == nil {
		ti.SetValue(filepath.Join(home, fmt.Sprintf("lfg-preset-%s.toml",
			time.Now().Format("2006-01-02"))))
	} else {
		ti.SetValue("lfg-preset.toml")
	}
	ti.Focus()
	return exportModel{palette: p, bundles: bundles, input: ti}
}

func (m exportModel) Init() tea.Cmd { return textinput.Blink }

func (m exportModel) Update(msg tea.Msg) (exportModel, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "esc", "delete":
			// Note: backspace is reserved for the textinput's delete-prev
			// when not yet saved. Use Esc / Delete to back out.
			return m, goTo(screenWelcome)
		case "enter":
			if m.saved {
				return m, goTo(screenWelcome)
			}
			path := strings.TrimSpace(m.input.Value())
			if path == "" {
				m.err = "enter a file path"
				return m, nil
			}
			if err := preset.Save(path, m.bundles); err != nil {
				m.err = err.Error()
				return m, nil
			}
			m.err = ""
			m.saved = true
			m.savedAt = path
			return m, nil
		}
	}

	if !m.saved {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m exportModel) View(width, height int) string {
	p := m.palette
	canvasW := width - 4
	if canvasW > 100 {
		canvasW = 100
	}
	if canvasW < 56 {
		canvasW = 56
	}
	contentW := canvasW - 4

	var b strings.Builder

	if m.saved {
		b.WriteString(SectionLabel(p, "Preset saved", "", contentW))
		b.WriteString("\n\n")
		check := lipgloss.NewStyle().Foreground(p.Success).Bold(true).Render("●")
		path := lipgloss.NewStyle().Foreground(p.Primary).Bold(true).Render(m.savedAt)
		b.WriteString("  " + check + "  saved to:\n")
		b.WriteString("       " + path + "\n\n")
		b.WriteString("  " + lipgloss.NewStyle().Foreground(p.Text).Bold(true).
			Render("What's next") + "\n")
		next := lipgloss.NewStyle().Foreground(p.Muted)
		b.WriteString("  " + next.Render(
			"• Move this file anywhere — it's just text, no secrets inside.") + "\n")
		b.WriteString("  " + next.Render(
			"• Use it on another machine:  lfg --config "+m.savedAt) + "\n")
	} else {
		b.WriteString(SectionLabel(p, "Save preset to TOML", "", contentW))
		b.WriteString("\n\n")
		desc := lipgloss.NewStyle().Foreground(p.Muted).
			Render("Captures every bundle + tool currently in lfg's preset (no dotfile content,\n  no secrets — just the install recipe). Edit the path below if you want.")
		b.WriteString("  " + desc + "\n\n")
		b.WriteString("  " + m.input.View() + "\n")
	}

	if m.err != "" {
		b.WriteString("\n  " + lipgloss.NewStyle().Foreground(p.Error).Bold(true).
			Render("✗ "+m.err) + "\n")
	}

	hint := HintLine(p,
		KeyHint(p, "⏎", "save"),
		KeyHint(p, "⎋", "back"),
	)
	if m.saved {
		hint = HintLine(p, KeyHint(p, "⏎", "back to menu"))
	}
	return Frame(p, width, height, "export preset", b.String(), hint, height < 22)
}
