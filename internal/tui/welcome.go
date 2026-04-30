package tui

import (
	"github.com/charmbracelet/huh"
	tea "github.com/charmbracelet/bubbletea"
)

// welcomeModel — first-screen picker driven by huh.Select.
type welcomeModel struct {
	palette Palette
	form    *huh.Form
	choice  string
}

const (
	welcomeInstall = "install"
	welcomeBackup  = "backup"
	welcomeQuit    = "quit"
)

func newWelcome(p Palette) welcomeModel {
	m := welcomeModel{palette: p}
	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("What would you like to do?").
				Description("Fresh machine detected. Pick your path.").
				Options(
					huh.NewOption("Install recommended setup", welcomeInstall),
					huh.NewOption("Backup this machine to a file", welcomeBackup),
					huh.NewOption("Quit", welcomeQuit),
				).
				Value(&m.choice),
		),
	).
		WithTheme(HuhTheme(p)).
		WithShowHelp(false).
		WithShowErrors(false)
	return m
}

func (m welcomeModel) Init() tea.Cmd { return m.form.Init() }

func (m welcomeModel) Update(msg tea.Msg) (welcomeModel, tea.Cmd) {
	f, cmd := m.form.Update(msg)
	if ff, ok := f.(*huh.Form); ok {
		m.form = ff
	}
	if m.form.State == huh.StateCompleted {
		switch m.choice {
		case welcomeInstall:
			return m, goTo(screenBundles)
		case welcomeBackup:
			return m, goTo(screenBackupPrompt)
		default:
			return m, goTo(screenQuit)
		}
	}
	return m, cmd
}

func (m welcomeModel) View(width, height int) string {
	return Frame(m.palette, width, height,
		"bootstrap your dev machine",
		m.form.View(),
		HintLine(m.palette,
			KeyHint(m.palette, "↑/↓", "move"),
			KeyHint(m.palette, "enter", "select"),
			KeyHint(m.palette, "ctrl-c", "quit"),
		),
		height < 22,
	)
}
