package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/anuj/lfg/internal/preset"
)

// confirmModel shows a summary panel + huh.Confirm (install/back).
type confirmModel struct {
	palette   Palette
	form      *huh.Form
	confirmed bool
	toInstall []preset.Tool
	alreadyOK int
	bySource  map[string]int
}

func newConfirm(p Palette, bundles []preset.Bundle, selected map[string]bool) confirmModel {
	m := confirmModel{palette: p, bySource: map[string]int{}}
	for _, b := range bundles {
		for _, t := range b.Tools {
			key := b.ID + "/" + t.Name
			if !selected[key] {
				if t.Installed {
					m.alreadyOK++
				}
				continue
			}
			m.toInstall = append(m.toInstall, t)
			m.bySource[t.Source]++
		}
	}

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Install %d tools?", len(m.toInstall))).
				Description("This is a UX prototype — no real changes yet.").
				Affirmative("Let's go").
				Negative("Back").
				Value(&m.confirmed),
		),
	).
		WithTheme(HuhTheme(p)).
		WithShowHelp(false)

	return m
}

func (m confirmModel) Init() tea.Cmd { return m.form.Init() }

func (m confirmModel) Update(msg tea.Msg) (confirmModel, tea.Cmd) {
	f, cmd := m.form.Update(msg)
	if ff, ok := f.(*huh.Form); ok {
		m.form = ff
	}
	if m.form.State == huh.StateCompleted {
		if m.confirmed {
			return m, goTo(screenProgress)
		}
		return m, goTo(screenTools)
	}
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "esc" {
		return m, goTo(screenTools)
	}
	return m, cmd
}

func (m confirmModel) View(width, height int) string {
	// Summary panel built from three little columns joined horizontally.
	stat := func(label, value string, col lipgloss.Color) string {
		v := lipgloss.NewStyle().Foreground(col).Bold(true).Render(value)
		l := lipgloss.NewStyle().Foreground(m.palette.Muted).Render(label)
		return lipgloss.JoinVertical(lipgloss.Center, v, l)
	}

	summary := lipgloss.JoinHorizontal(lipgloss.Center,
		lipgloss.NewStyle().Padding(0, 3).Render(stat("to install", fmt.Sprintf("%d", len(m.toInstall)), m.palette.Primary)),
		lipgloss.NewStyle().Padding(0, 3).Render(stat("already ok", fmt.Sprintf("%d", m.alreadyOK), m.palette.Success)),
		lipgloss.NewStyle().Padding(0, 3).Render(stat("sources", fmt.Sprintf("%d", len(m.bySource)), m.palette.Accent)),
	)

	summaryBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.palette.Subtle).
		Padding(1, 2).
		Render(summary)

	// Source breakdown line
	srcLine := renderSources(m.palette, m.bySource)

	inner := lipgloss.JoinVertical(lipgloss.Center,
		summaryBox,
		"",
		srcLine,
		"",
		m.form.View(),
	)

	return Frame(m.palette, width, height,
		"step 3 of 3  ·  ready",
		inner,
		HintLine(m.palette,
			KeyHint(m.palette, "enter", "install"),
			KeyHint(m.palette, "esc", "back"),
		),
		height < 22,
	)
}

func renderSources(p Palette, m map[string]int) string {
	if len(m) == 0 {
		return ""
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		kStyle := lipgloss.NewStyle().Foreground(p.Accent).Bold(true).Render(k)
		vStyle := lipgloss.NewStyle().Foreground(p.Text).Render(fmt.Sprintf("%d", m[k]))
		parts = append(parts, kStyle+" "+vStyle)
	}
	sep := lipgloss.NewStyle().Foreground(p.Subtle).Render(" · ")
	return lipgloss.NewStyle().Foreground(p.Muted).Render("via ") + strings.Join(parts, sep)
}
