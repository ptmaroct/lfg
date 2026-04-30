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

// progressModel simulates an install run. Spinner + gradient progress + log tail.
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

	// lipgloss v1 / bubbles v1.0 progress uses gradient constructors.
	// Set the gradient to our palette extremes.
	bar := progress.New(
		progress.WithGradient(string(p.Primary), string(p.Success)),
		progress.WithWidth(44),
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
		check := lipgloss.NewStyle().Foreground(m.palette.Success).Render("✓")
		src := lipgloss.NewStyle().Foreground(m.palette.Muted).Render("(" + t.Source + ")")
		m.logs = append(m.logs, fmt.Sprintf("%s installed %s %s", check, t.Name, src))
		if len(m.logs) > 8 {
			m.logs = m.logs[len(m.logs)-8:]
		}
		m.index = msg.idx + 1
		return m, m.schedule(m.index)
	case progress.FrameMsg:
		pm, cmd := m.bar.Update(msg)
		m.bar = pm.(progress.Model)
		return m, cmd
	}
	return m, nil
}

func (m progressModel) View(width, height int) string {
	pct := 0.0
	if len(m.queue) > 0 {
		pct = float64(m.index) / float64(len(m.queue))
	}

	// Stats row: N/total  ·  elapsed  ·  eta
	numStyle := lipgloss.NewStyle().Foreground(m.palette.Primary).Bold(true)
	muStyle := lipgloss.NewStyle().Foreground(m.palette.Muted)
	stats := lipgloss.JoinHorizontal(lipgloss.Center,
		muStyle.Render("progress "),
		numStyle.Render(fmt.Sprintf("%d/%d", m.index, len(m.queue))),
		muStyle.Render("  ·  elapsed "),
		numStyle.Render(m.stopwatch.View()),
	)

	current := lipgloss.NewStyle().Foreground(m.palette.Muted).Render("— done —")
	if m.index < len(m.queue) {
		bold := lipgloss.NewStyle().Foreground(m.palette.Text).Bold(true)
		src := lipgloss.NewStyle().Foreground(m.palette.Muted)
		current = fmt.Sprintf("%s installing %s %s",
			m.spinner.View(),
			bold.Render(m.queue[m.index].Name),
			src.Render("("+m.queue[m.index].Source+")"),
		)
	} else if m.done {
		current = lipgloss.NewStyle().Foreground(m.palette.Success).Bold(true).Render("✓ all tools installed")
	}

	logBox := renderLogs(m.palette, m.logs)

	inner := lipgloss.JoinVertical(lipgloss.Center,
		stats,
		"",
		m.bar.ViewAs(pct),
		"",
		current,
		"",
		logBox,
	)

	return Frame(m.palette, width, height,
		"installing",
		inner,
		HintLine(m.palette,
			KeyHint(m.palette, "ctrl-c", "cancel"),
		),
		height < 22,
	)
}

func renderLogs(p Palette, lines []string) string {
	if len(lines) == 0 {
		return lipgloss.NewStyle().Foreground(p.Muted).Italic(true).Render("(log will appear here)")
	}
	content := strings.Join(lines, "\n")
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.Subtle).
		Foreground(p.Text).
		Padding(0, 2).
		Width(56).
		Render(content)
}
