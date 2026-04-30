package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// backupModel — encrypt y/n prompt (huh.NewConfirm) → fake pack
// (bubbles.Spinner) → result card.
type backupModel struct {
	palette  Palette
	step     int // 0=prompt, 1=running, 2=done
	form     *huh.Form
	encrypt  bool
	spinner  spinner.Model
	filepath string
}

type backupDoneMsg struct{}

func newBackup(p Palette) backupModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(p.Primary)
	m := backupModel{palette: p, spinner: sp, encrypt: true}
	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Encrypt the backup?").
				Description("Encrypted = .tar.age (age). Plain = .tar.gz, no secrets.").
				Affirmative(".tar.age").
				Negative(".tar.gz").
				Value(&m.encrypt),
		),
	).
		WithTheme(HuhTheme(p)).
		WithShowHelp(false).
		WithShowErrors(false)
	return m
}

func (m backupModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.form.Init())
}

func (m backupModel) Update(msg tea.Msg) (backupModel, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case backupDoneMsg:
		ts := time.Now().Format("2006-01-02")
		if m.encrypt {
			m.filepath = "~/lfg-backup-" + ts + ".tar.age"
		} else {
			m.filepath = "~/lfg-backup-" + ts + ".tar.gz"
		}
		m.step = 2
		return m, nil
	case tea.KeyMsg:
		switch m.step {
		case 0:
			if msg.String() == "esc" {
				return m, goTo(screenWelcome)
			}
		case 2:
			if msg.String() != "q" { // global quit handler owns `q`
				return m, tea.Quit
			}
		}
	}

	if m.step == 0 {
		f, cmd := m.form.Update(msg)
		if ff, ok := f.(*huh.Form); ok {
			m.form = ff
		}
		if m.form.State == huh.StateCompleted {
			m.step = 1
			return m, tea.Tick(1200*time.Millisecond, func(time.Time) tea.Msg {
				return backupDoneMsg{}
			})
		}
		return m, cmd
	}
	return m, nil
}

func (m backupModel) View(width, height int) string {
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

	switch m.step {
	case 0:
		b.WriteString(m.form.View())
	case 1:
		b.WriteString(SectionLabel(p, "Packing", "", contentW))
		b.WriteString("\n\n")
		label := lipgloss.NewStyle().Foreground(p.Text).Render("packing your machine...")
		sub := lipgloss.NewStyle().Foreground(p.Muted).Italic(true).Render("dotfiles · brew list · configs")
		b.WriteString("  " + m.spinner.View() + "  " + label + "\n\n")
		b.WriteString("  " + sub)
	case 2:
		b.WriteString(SectionLabel(p, "Backup written", "", contentW))
		b.WriteString("\n\n")
		check := lipgloss.NewStyle().Foreground(p.Success).Bold(true).Render("●")
		path := lipgloss.NewStyle().
			Foreground(p.Primary).
			Bold(true).
			Border(lipgloss.NormalBorder()).
			BorderForeground(p.Subtle).
			Padding(0, 2).
			Render(m.filepath)
		b.WriteString("  " + check + "  saved\n\n")
		b.WriteString("  " + path + "\n")
		if m.encrypt {
			b.WriteString("\n")
			warn := lipgloss.NewStyle().Foreground(p.Warn).Bold(true).Render("⚠ ")
			note := lipgloss.NewStyle().Foreground(p.Muted).Italic(true).
				Render("back up ~/.config/lfg/key.txt separately — without it, archive is unrecoverable")
			b.WriteString("  " + warn + note)
		}
	}

	hint := HintLine(p,
		KeyHint(p, "←→", "switch"),
		KeyHint(p, "⏎", "confirm"),
		KeyHint(p, "⎋", "back"),
	)
	if m.step == 2 {
		hint = HintLine(p, KeyHint(p, "any", "exit"))
	}

	return Frame(p, width, height, "backup", b.String(), hint, height < 22)
}
