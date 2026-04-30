package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// welcomeModel — first-screen entry. Custom render (not huh) so we can
// place the figlet hero AND the numbered action list in one composition.
type welcomeModel struct {
	palette Palette
	cursor  int
	choices []welcomeChoice
}

type welcomeChoice struct {
	label  string
	desc   string
	action screen
}

func newWelcome(p Palette) welcomeModel {
	return welcomeModel{
		palette: p,
		choices: []welcomeChoice{
			{
				label:  "INSTALL RECOMMENDED SETUP",
				desc:   "Pick bundles, install tools. Opinionated but customizable.",
				action: screenTree,
			},
			{
				label:  "BACKUP THIS MACHINE",
				desc:   "Snapshot dotfiles, package list, configs into a single file.",
				action: screenBackupPrompt,
			},
			{
				label:  "QUIT",
				desc:   "Exit lfg.",
				action: screenQuit,
			},
		},
	}
}

func (m welcomeModel) Init() tea.Cmd { return nil }

func (m welcomeModel) Update(msg tea.Msg) (welcomeModel, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			return m, goTo(m.choices[m.cursor].action)
		case "esc", "q":
			return m, tea.Quit
		case "1":
			m.cursor = 0
			return m, goTo(m.choices[0].action)
		case "2":
			m.cursor = 1
			return m, goTo(m.choices[1].action)
		case "3":
			m.cursor = 2
			return m, goTo(m.choices[2].action)
		}
	}
	return m, nil
}

func (m welcomeModel) View(width, height int) string {
	p := m.palette

	canvasW := width - 4
	if canvasW > 100 {
		canvasW = 100
	}
	if canvasW < 56 {
		canvasW = 56
	}
	contentW := canvasW - 4
	compact := height < 22

	var b strings.Builder

	// Hero block — figlet on big terminals, single-line gradient on small.
	hero := RenderTitle(p, "lfg", compact)
	b.WriteString(lipgloss.PlaceHorizontal(contentW, lipgloss.Center, hero))
	b.WriteString("\n")
	tagline := lipgloss.NewStyle().Foreground(p.Muted).Italic(true).
		Render("a new dev machine, in less time than this hint takes to read")
	b.WriteString(lipgloss.PlaceHorizontal(contentW, lipgloss.Center, tagline))
	b.WriteString("\n\n")

	// Actions section
	b.WriteString(SectionLabel(p, "What now?", "fresh machine detected", contentW))
	b.WriteString("\n\n")

	for i, c := range m.choices {
		b.WriteString(m.renderChoice(i, c, contentW))
		b.WriteString("\n")
	}

	return Frame(p, width, height,
		"welcome",
		b.String(),
		HintLine(p,
			KeyHint(p, "↑↓", "nav"),
			KeyHint(p, "1-3", "jump"),
			KeyHint(p, "⏎", "select"),
			KeyHint(p, "⎋", "quit"),
		),
		compact,
	)
}

func (m welcomeModel) renderChoice(i int, c welcomeChoice, contentW int) string {
	p := m.palette
	num := lipgloss.NewStyle().Foreground(p.Muted).Render(strings.Repeat("0", 1) + sprint1(i+1))
	gutter := "  "
	titleStyle := lipgloss.NewStyle().Foreground(p.Text).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(p.Muted).Italic(true)

	if i == m.cursor {
		gutter = lipgloss.NewStyle().Foreground(p.Primary).Bold(true).Render("▸ ")
		num = lipgloss.NewStyle().Foreground(p.Primary).Bold(true).Render("0" + sprint1(i+1))
		titleStyle = titleStyle.Foreground(p.Primary)
	}

	line1 := gutter + num + "  " + titleStyle.Render(c.label)
	line2 := "       " + descStyle.Render(c.desc)
	_ = contentW
	return line1 + "\n" + line2
}

// sprint1 — tiny int→string for 1..9. Avoids importing strconv.
func sprint1(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return ""
}
