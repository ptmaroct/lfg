package tui

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/stopwatch"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/anuj/lfg/internal/preset"
)

// progressModel — instrument-panel install screen. Stat row top, gradient
// progress bar, current-action line, log tail. Mocked installs sleep 300-900ms
// per tool so the UX feels alive without doing real work.
type progressModel struct {
	palette   Palette
	queue     []preset.Tool
	index     int
	logs      []string
	spinner   spinner.Model
	bar       progress.Model
	stopwatch stopwatch.Model
	done      bool
}

type installStepMsg struct{ idx int }

func newProgress(p Palette, bundles []preset.Bundle, selected map[string]bool) progressModel {
	var queue []preset.Tool
	for _, b := range bundles {
		for _, t := range b.Tools {
			if selected[b.ID+"/"+t.Name] {
				queue = append(queue, t)
			}
		}
	}

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(p.Primary)

	bar := progress.New(
		progress.WithGradient(string(p.Primary), string(p.Success)),
		progress.WithWidth(60),
		progress.WithoutPercentage(),
	)

	sw := stopwatch.NewWithInterval(100 * time.Millisecond)

	return progressModel{
		palette:   p,
		queue:     queue,
		spinner:   sp,
		bar:       bar,
		stopwatch: sw,
	}
}

func (m progressModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.stopwatch.Start(), m.schedule(0))
}

func (m progressModel) schedule(idx int) tea.Cmd {
	return tea.Tick(time.Duration(300+rand.Intn(600))*time.Millisecond, func(time.Time) tea.Msg {
		return installStepMsg{idx: idx}
	})
}

func (m progressModel) Update(msg tea.Msg) (progressModel, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case stopwatch.TickMsg, stopwatch.StartStopMsg, stopwatch.ResetMsg:
		var cmd tea.Cmd
		m.stopwatch, cmd = m.stopwatch.Update(msg)
		return m, cmd
	case installStepMsg:
		if msg.idx >= len(m.queue) {
			m.done = true
			return m, tea.Batch(
				m.stopwatch.Stop(),
				tea.Tick(800*time.Millisecond, func(time.Time) tea.Msg {
					return transitionMsg{target: screenDone}
				}),
			)
		}
		t := m.queue[msg.idx]
		check := lipgloss.NewStyle().Foreground(m.palette.Success).Render("●")
		name := lipgloss.NewStyle().Foreground(m.palette.Text).Render(t.Name)
		src := lipgloss.NewStyle().Foreground(m.palette.Muted).Render(t.Source)
		m.logs = append(m.logs, fmt.Sprintf("%s  %-26s  %s", check, name, src))
		if len(m.logs) > 8 {
			m.logs = m.logs[len(m.logs)-8:]
		}
		m.index = msg.idx + 1
		return m, m.schedule(m.index)
	case progress.FrameMsg:
		pm, cmd := m.bar.Update(msg)
		m.bar = pm.(progress.Model)
		return m, cmd
	case tea.KeyMsg:
		if msg.String() == "esc" {
			return m, goTo(screenConfirm)
		}
	}
	return m, nil
}

func (m progressModel) View(width, height int) string {
	p := m.palette
	canvasW := width - 4
	if canvasW > 100 {
		canvasW = 100
	}
	if canvasW < 56 {
		canvasW = 56
	}
	contentW := canvasW - 4

	pct := 0.0
	if len(m.queue) > 0 {
		pct = float64(m.index) / float64(len(m.queue))
	}

	var b strings.Builder
	b.WriteString(SectionLabel(p, "Installing", fmt.Sprintf("%d of %d", m.index, len(m.queue)), contentW))
	b.WriteString("\n\n")

	// Big stat row
	b.WriteString(renderStatRow(p, contentW, []statCell{
		{label: "DONE", value: fmt.Sprintf("%02d", m.index), color: p.Primary},
		{label: "TOTAL", value: fmt.Sprintf("%02d", len(m.queue)), color: p.Text},
		{label: "ELAPSED", value: m.stopwatch.View(), color: p.Accent},
		{label: "PCT", value: fmt.Sprintf("%d%%", int(pct*100)), color: p.Success},
	}))
	b.WriteString("\n\n")

	// Bar (full width)
	b.WriteString("  ")
	b.WriteString(m.bar.ViewAs(pct))
	b.WriteString("\n\n")

	// Current action
	current := lipgloss.NewStyle().Foreground(p.Muted).Render("— done —")
	if m.index < len(m.queue) {
		bold := lipgloss.NewStyle().Foreground(p.Primary).Bold(true)
		mut := lipgloss.NewStyle().Foreground(p.Muted)
		current = fmt.Sprintf("%s %s %s",
			m.spinner.View(),
			bold.Render(m.queue[m.index].Name),
			mut.Render("· "+m.queue[m.index].Source),
		)
	} else if m.done {
		current = lipgloss.NewStyle().Foreground(p.Success).Bold(true).Render("● all installed")
	}
	b.WriteString("  " + current + "\n\n")

	// Log tail
	b.WriteString("  " + Hairline(p, contentW-2) + "\n")
	b.WriteString("  " + lipgloss.NewStyle().Foreground(p.Muted).Render("LOG") + "\n")
	if len(m.logs) == 0 {
		b.WriteString("  " + lipgloss.NewStyle().Foreground(p.Muted).Italic(true).Render("(empty)") + "\n")
	}
	for _, line := range m.logs {
		b.WriteString("  " + line + "\n")
	}
	b.WriteString("  " + Hairline(p, contentW-2))

	return Frame(p, width, height,
		"installing",
		b.String(),
		HintLine(p,
			KeyHint(p, "⎋", "cancel"),
			KeyHint(p, "^C", "quit"),
		),
		height < 22,
	)
}
