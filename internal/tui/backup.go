package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"

	"github.com/ptmaroct/lfg/internal/backup"
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

// renderBackupPreview lists the backup sources in plain language with a
// presence dot (● = found on disk, ○ = nothing here). Lets the user
// see exactly which files lfg will copy before they confirm.
func renderBackupPreview(p Palette) string {
	type item struct {
		label  string
		paths  []string
		exists bool
	}
	home, _ := os.UserHomeDir()
	exists := func(rel string) bool {
		_, err := os.Lstat(filepath.Join(home, rel))
		return err == nil
	}
	any := func(rels ...string) bool {
		for _, r := range rels {
			if exists(r) {
				return true
			}
		}
		return false
	}

	groups := []item{
		{label: "Shell config", paths: []string{".zshrc", ".zprofile", ".zshenv", ".bashrc", ".bash_profile", ".profile"}},
		{label: "Git, tmux, vim, editorconfig", paths: []string{".gitconfig", ".tmux.conf", ".vimrc", ".editorconfig", ".inputrc"}},
		{label: "Starship + dev tool configs", paths: []string{".config/starship.toml", ".config/starship", ".config/mise", ".config/bat", ".config/btop", ".config/lazygit", ".config/yazi", ".config/glow"}},
		{label: "Editor configs (nvim, zed, ghostty)", paths: []string{".config/nvim", ".config/zed", ".config/ghostty", ".config/zellij"}},
		{label: "Claude Code + Codex settings", paths: []string{".claude/settings.json", ".claude/CLAUDE.md", ".claude/agents", ".claude/commands", ".codex/config.toml"}},
		{label: "SSH config + public keys (private keys NEVER copied)", paths: []string{".ssh"}},
	}
	for i := range groups {
		groups[i].exists = any(groups[i].paths...)
	}

	muted := lipgloss.NewStyle().Foreground(p.Muted)
	text := lipgloss.NewStyle().Foreground(p.Text)
	successDot := lipgloss.NewStyle().Foreground(p.Success).Render("●")
	missDot := lipgloss.NewStyle().Foreground(p.Subtle).Render("○")

	var b strings.Builder
	for _, g := range groups {
		dot := missDot
		label := muted.Render(g.label) + lipgloss.NewStyle().Foreground(p.Subtle).Italic(true).Render("  (nothing here)")
		if g.exists {
			dot = successDot
			label = text.Render(g.label)
		}
		b.WriteString("  " + dot + "  " + label + "\n")
	}
	return b.String()
}

func newBackup(p Palette) backupModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(p.Primary)
	m := backupModel{palette: p, spinner: sp, encrypt: true}
	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Should we lock this backup with a key?").
				Description(
					"Your backup includes shell history, env files, and SSH config — anything\n"+
						"you wouldn't want a stranger reading. Locking it scrambles the file with a\n"+
						"key we save to ~/.config/lfg/key.txt. Keep that key safe (a password\n"+
						"manager works) — without it the backup can't be opened, not even by you.\n"+
						"\n"+
						"Pick \"Lock it\" if the backup might leave your machine (cloud, USB, email).\n"+
						"Pick \"Skip — local only\" if it stays on this machine and you want to skim it.",
				).
				Affirmative("Lock it (recommended)").
				Negative("Skip — local only").
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
			switch msg.String() {
			case "esc", "backspace", "delete":
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
		// Show what's actually going into the backup so the user sees
		// the scope before answering the lock-it-or-not question. Big
		// driver of the "what does this even back up?" complaint.
		b.WriteString(SectionLabel(p, "What we'll back up", "from your home folder", contentW))
		b.WriteString("\n")
		b.WriteString("  " + lipgloss.NewStyle().Foreground(p.Muted).Italic(true).
			Render("Anything missing on your machine is silently skipped — nothing is required.") + "\n\n")
		b.WriteString(renderBackupPreview(p) + "\n\n")
		b.WriteString(m.form.View())
	case 1:
		b.WriteString(SectionLabel(p, "Packing your machine", "", contentW))
		b.WriteString("\n\n")
		label := lipgloss.NewStyle().Foreground(p.Text).Render("collecting dotfiles, brew list, configs...")
		sub := lipgloss.NewStyle().Foreground(p.Muted).Italic(true).
			Render("this usually takes a few seconds")
		b.WriteString("  " + m.spinner.View() + "  " + label + "\n\n")
		b.WriteString("  " + sub)
	case 2:
		b.WriteString(SectionLabel(p, "Backup ready", "", contentW))
		b.WriteString("\n\n")
		check := lipgloss.NewStyle().Foreground(p.Success).Bold(true).Render("●")
		path := lipgloss.NewStyle().Foreground(p.Primary).Bold(true).Render(m.filepath)
		b.WriteString("  " + check + "  saved to:\n")
		b.WriteString("       " + path + "\n\n")

		stats := fmt.Sprintf("%d files · %s",
			m.result.Files, humanize.Bytes(uint64(m.result.Bytes)))
		if m.result.Skipped > 0 || m.result.Excluded > 0 {
			stats += fmt.Sprintf(" (skipped %d, excluded %d)",
				m.result.Skipped, m.result.Excluded)
		}
		b.WriteString("  " + lipgloss.NewStyle().Foreground(p.Muted).Render(stats) + "\n\n")

		// Plain-language "what now" so the user actually knows what to
		// do with the file. Tailored to encrypted vs plain because the
		// next steps differ a lot.
		next := lipgloss.NewStyle().Foreground(p.Muted)
		if m.encrypt {
			b.WriteString("  " + lipgloss.NewStyle().Foreground(p.Text).Bold(true).
				Render("What's next") + "\n")
			b.WriteString("  " + next.Render(
				"• Move this file anywhere safe — cloud, external drive, etc.") + "\n")
			b.WriteString("  " + next.Render(
				"• To open it later, run:  lfg backup --restore "+m.filepath) + "\n\n")

			warn := lipgloss.NewStyle().Foreground(p.Warn).Bold(true).Render("⚠  ")
			keyMsg := "Your unlock key lives at " + m.result.KeyPath + ".\n     Back it up too — without it nobody can open this file."
			if m.result.NewKey {
				keyMsg = "We just created a new unlock key at " + m.result.KeyPath + ".\n     Save it RIGHT NOW (1Password, secure note) — without it nothing here can be opened later."
			}
			note := lipgloss.NewStyle().Foreground(p.Text).Render(keyMsg)
			b.WriteString("  " + warn + note)
		} else {
			b.WriteString("  " + lipgloss.NewStyle().Foreground(p.Text).Bold(true).
				Render("What's next") + "\n")
			b.WriteString("  " + next.Render(
				"• This file is NOT password-protected — keep it on this machine only.") + "\n")
			b.WriteString("  " + next.Render(
				"• Peek inside any time:  tar -tzf "+m.filepath) + "\n")
			b.WriteString("  " + next.Render(
				"• Restore later:  lfg backup --restore "+m.filepath))
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
