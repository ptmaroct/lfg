package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	"github.com/ptmaroct/lfg/internal/preset"
)

// configInputModel — modal for "load config file from welcome screen".
// Single text input that accepts either a local path or http(s) URL.
// On submit it calls preset.Load synchronously; on success it transitions
// to screenProbe carrying the loaded bundles via transitionMsg.bundles.
// Errors stay rendered below the input until the user fixes the path.
type configInputModel struct {
	palette Palette
	input   textinput.Model
	err     string
}

func newConfigInput(p Palette) configInputModel {
	ti := textinput.New()
	ti.Placeholder = "/path/to/preset.toml or https://example.com/preset.toml"
	ti.Prompt = "› "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(p.Primary).Bold(true)
	ti.TextStyle = lipgloss.NewStyle().Foreground(p.Text)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(p.Muted).Italic(true)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(p.Primary)
	ti.CharLimit = 1024
	ti.Width = 80
	ti.Focus()
	return configInputModel{palette: p, input: ti}
}

func (m configInputModel) Init() tea.Cmd { return textinput.Blink }

func (m configInputModel) Update(msg tea.Msg) (configInputModel, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "esc", "delete":
			// backspace stays on the textinput for editing.
			return m, goTo(screenWelcome)
		case "enter":
			spec := strings.TrimSpace(m.input.Value())
			if spec == "" {
				m.err = "enter a path or URL"
				return m, nil
			}
			loaded, err := preset.Load(spec)
			if err != nil {
				m.err = err.Error()
				return m, nil
			}
			m.err = ""
			groups := preset.GroupAliases(preset.FilterAliasesForHost(loaded.Aliases))
			return m, func() tea.Msg {
				return transitionMsg{
					target:         screenProbe,
					bundles:        preset.FilterForHost(loaded.Bundles),
					aliasGroups:    groups,
					replaceAliases: true,
				}
			}
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m configInputModel) View(width, height int) string {
	p := m.palette
	canvasW := CanvasW(width)
	contentW := canvasW - 4

	var b strings.Builder
	b.WriteString(SectionLabel(p, "Load preset", "local file or http(s) URL", contentW))
	b.WriteString("\n\n")
	b.WriteString(lipgloss.NewStyle().Foreground(p.Muted).
		Render("  Paste a TOML preset path. We'll fetch + parse it, then run detect."))
	b.WriteString("\n\n")
	b.WriteString("  " + m.input.View() + "\n")

	if m.err != "" {
		b.WriteString("\n  " + lipgloss.NewStyle().Foreground(p.Error).Bold(true).
			Render("✗ "+m.err) + "\n")
	}

	return Frame(p, width, height,
		"load config",
		b.String(),
		HintLine(p,
			KeyHint(p, "⏎", "load"),
			KeyHint(p, "ESC", "back"),
		),
		height < 22,
	)
}
