package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"

	"github.com/anuj/lfg/internal/backup"
)

// backupModel — encrypt y/n prompt (huh.NewConfirm) → real pack via
// internal/backup.Pack (bubbles.Spinner during the goroutine) → result
// card with file path + counters.
type backupModel struct {
	palette  Palette
	step     int // 0=prompt, 1=running, 2=done, 3=error
	form     *huh.Form
	encrypt  bool
	spinner  spinner.Model
	filepath string
	result   backup.Result
	err      error
}

type backupDoneMsg struct {
	result backup.Result
	err    error
}

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
		m.result = msg.result
		m.err = msg.err
		m.filepath = msg.result.Path
		if msg.err != nil {
			m.step = 3
		} else {
			m.step = 2
		}
		return m, nil
	case tea.KeyMsg:
		switch m.step {
		case 0:
			if msg.String() == "esc" {
				return m, goTo(screenWelcome)
			}
		case 2, 3:
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
			encrypt := m.encrypt
			return m, func() tea.Msg {
				home, _ := os.UserHomeDir()
				r, err := backup.Pack(backup.Options{
					OutDir:  home,
					Encrypt: encrypt,
				})
				return backupDoneMsg{result: r, err: err}
			}
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
		b.WriteString("  " + path + "\n\n")

		stats := fmt.Sprintf("%d files · %s · %d skipped · %d excluded",
			m.result.Files, humanize.Bytes(uint64(m.result.Bytes)),
			m.result.Skipped, m.result.Excluded)
		b.WriteString("  " + lipgloss.NewStyle().Foreground(p.Muted).Render(stats) + "\n")

		if m.encrypt {
			b.WriteString("\n")
			warn := lipgloss.NewStyle().Foreground(p.Warn).Bold(true).Render("⚠ ")
			msg := "back up " + m.result.KeyPath + " separately"
			if m.result.NewKey {
				msg = "GENERATED new key at " + m.result.KeyPath + " — back it up NOW"
			}
			note := lipgloss.NewStyle().Foreground(p.Muted).Italic(true).Render(msg)
			b.WriteString("  " + warn + note)
		}
	case 3:
		b.WriteString(SectionLabel(p, "Backup failed", "", contentW))
		b.WriteString("\n\n")
		x := lipgloss.NewStyle().Foreground(p.Warn).Bold(true).Render("✗")
		b.WriteString("  " + x + "  ")
		b.WriteString(lipgloss.NewStyle().Foreground(p.Text).Render(m.err.Error()))
	}

	hint := HintLine(p,
		KeyHint(p, "←→", "switch"),
		KeyHint(p, "⏎", "confirm"),
		KeyHint(p, "⎋", "back"),
	)
	if m.step == 2 || m.step == 3 {
		hint = HintLine(p, KeyHint(p, "any", "exit"))
	}

	return Frame(p, width, height, "backup", b.String(), hint, height < 22)
}
