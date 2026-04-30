package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// backupModel — encrypt y/n prompt → fake pack → result card.
type backupModel struct {
	palette  Palette
	step     int // 0=prompt, 1=running, 2=done
	encrypt  bool
	cursor   int // 0=encrypt, 1=plain
	spinner  spinner.Model
	filepath string
}

type backupDoneMsg struct{}

func newBackup(p Palette) backupModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(p.Primary)
	return backupModel{palette: p, spinner: sp}
}

func (m backupModel) Init() tea.Cmd { return m.spinner.Tick }

func (m backupModel) Update(msg tea.Msg) (backupModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.step {
		case 0:
			switch msg.String() {
			case "left", "h", "right", "l", "tab":
				m.cursor = 1 - m.cursor
			case "enter":
				m.encrypt = m.cursor == 0
				m.step = 1
				return m, tea.Tick(1200*time.Millisecond, func(time.Time) tea.Msg {
					return backupDoneMsg{}
				})
			case "y", "Y":
				m.encrypt = true
				m.cursor = 0
				m.step = 1
				return m, tea.Tick(1200*time.Millisecond, func(time.Time) tea.Msg {
					return backupDoneMsg{}
				})
			case "n", "N":
				m.encrypt = false
				m.cursor = 1
				m.step = 1
				return m, tea.Tick(1200*time.Millisecond, func(time.Time) tea.Msg {
					return backupDoneMsg{}
				})
			case "esc":
				return m, goTo(screenWelcome)
			}
		case 2:
			return m, tea.Quit
		}
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
		b.WriteString(SectionLabel(p, "Encrypt the backup?", "", contentW))
		b.WriteString("\n\n")

		opts := []struct{ label, desc string }{
			{"YES   .tar.age", "Encrypted via age. Need ~/.config/lfg/key.txt to restore. Safer for dotfiles + env."},
			{"NO    .tar.gz", "Plain tar.gz. Inspectable. DO NOT include secrets."},
		}
		for i, o := range opts {
			num := lipgloss.NewStyle().Foreground(p.Muted).Render("0" + sprint1(i+1))
			gutter := "  "
			labelStyle := lipgloss.NewStyle().Foreground(p.Text).Bold(true)
			descStyle := lipgloss.NewStyle().Foreground(p.Muted).Italic(true)
			if i == m.cursor {
				gutter = lipgloss.NewStyle().Foreground(p.Primary).Bold(true).Render("▸ ")
				num = lipgloss.NewStyle().Foreground(p.Primary).Bold(true).Render("0" + sprint1(i+1))
				labelStyle = labelStyle.Foreground(p.Primary)
			}
			b.WriteString(gutter + num + "  " + labelStyle.Render(o.label) + "\n")
			b.WriteString("       " + descStyle.Render(o.desc) + "\n\n")
		}
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
		KeyHint(p, "Y/N", "pick"),
		KeyHint(p, "⏎", "confirm"),
		KeyHint(p, "⎋", "back"),
	)
	if m.step == 2 {
		hint = HintLine(p, KeyHint(p, "any", "exit"))
	}

	return Frame(p, width, height, "backup", b.String(), hint, height < 22)
}
