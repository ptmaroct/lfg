package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// backupModel — encrypt y/n → fake pack → result card.
type backupModel struct {
	palette  Palette
	step     int // 0=prompt, 1=running, 2=done
	encrypt  bool
	form     *huh.Form
	spinner  spinner.Model
	filepath string
}

type backupDoneMsg struct{}

func newBackup(p Palette) backupModel {
	m := backupModel{palette: p}
	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Encrypt this backup?").
				Description("Encrypted (.tar.age) needs ~/.config/lfg/key.txt to restore. Safer for dotfiles + env vars.").
				Affirmative("Encrypt").
				Negative("Plain tar").
				Value(&m.encrypt),
		),
	).
		WithTheme(HuhTheme(p)).
		WithShowHelp(false)

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(p.Primary)
	m.spinner = sp
	return m
}

func (m backupModel) Init() tea.Cmd {
	return tea.Batch(m.form.Init(), m.spinner.Tick)
}

func (m backupModel) Update(msg tea.Msg) (backupModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.step == 2 {
			return m, tea.Quit
		}
		if msg.String() == "esc" && m.step == 0 {
			return m, goTo(screenWelcome)
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
		return m, nil
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
	var inner string

	switch m.step {
	case 0:
		inner = m.form.View()
	case 1:
		label := lipgloss.NewStyle().Foreground(p.Text).Render("packing your machine...")
		sub := lipgloss.NewStyle().Foreground(p.Muted).Italic(true).Render("dotfiles · brew list · configs")
		inner = lipgloss.JoinVertical(lipgloss.Center,
			m.spinner.View()+"  "+label,
			"",
			sub,
		)
	case 2:
		check := lipgloss.NewStyle().Foreground(p.Success).Bold(true).Render("✓")
		title := lipgloss.NewStyle().Foreground(p.Text).Bold(true).Render("backup written")
		path := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(p.Accent).
			Foreground(p.Primary).
			Bold(true).
			Padding(0, 2).
			Render(m.filepath)
		parts := []string{
			check + "  " + title,
			"",
			path,
		}
		if m.encrypt {
			warn := lipgloss.NewStyle().Foreground(p.Warn).Bold(true).Render("⚠  back up ~/.config/lfg/key.txt separately")
			note := lipgloss.NewStyle().Foreground(p.Muted).Italic(true).Render("without it, this archive is unrecoverable")
			parts = append(parts, "", warn, note)
		}
		inner = lipgloss.JoinVertical(lipgloss.Center, parts...)
	}

	hint := HintLine(p,
		KeyHint(p, "enter", "confirm"),
		KeyHint(p, "esc", "back"),
	)
	if m.step == 2 {
		hint = HintLine(p, KeyHint(p, "any key", "exit"))
	}

	return Frame(p, width, height, "backup", inner, hint, height < 22)
}
