package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ptmaroct/lfg/internal/detect"
	"github.com/ptmaroct/lfg/internal/installer"
	"github.com/ptmaroct/lfg/internal/preset"
)

// probeModel — first-paint screen that runs detect.ProbeAll while
// streaming live counts to the user. Replaces the previously-silent
// pre-TUI probe in cli/default.go that made the terminal look frozen on
// slow systems.
//
// Lifecycle:
//
//  1. Init kicks off ProbeAllStream in a goroutine + arms the spinner +
//     a "wait next probe step" cmd.
//  2. Each probeStepMsg increments `done` and appends a log line.
//  3. When the channel closes (probeDoneMsg), we apply the results onto
//     the raw bundles, set the harness slice on the installer package,
//     and emit a transitionMsg carrying the probed bundles to welcome.
type probeModel struct {
	palette Palette
	bundles []preset.Bundle // raw, pre-probe
	spinner spinner.Model
	bar     progress.Model
	total   int
	done    int
	current string
	results map[string]detect.Result

	stream chan detect.ProbeStep
}

type probeStepMsg detect.ProbeStep
type probeDoneMsg struct {
	bundles []preset.Bundle
}

func newProbe(p Palette, bundles []preset.Bundle) probeModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(p.Primary)

	bar := progress.New(
		progress.WithGradient(hex(p.Primary), hex(p.Success)),
		progress.WithWidth(60),
		progress.WithoutPercentage(),
	)

	total := 0
	for _, b := range bundles {
		total += len(b.Tools)
	}

	return probeModel{
		palette: p,
		bundles: bundles,
		spinner: sp,
		bar:     bar,
		total:   total,
		results: map[string]detect.Result{},
		stream:  make(chan detect.ProbeStep, max2(total, 1)),
	}
}

func (m probeModel) Init() tea.Cmd {
	go detect.ProbeAllStream(m.bundles, m.stream)
	return tea.Batch(m.spinner.Tick, m.waitStep())
}

// waitStep returns a cmd that pulls one ProbeStep off the channel. When
// the channel closes, it applies results + emits probeDoneMsg.
func (m probeModel) waitStep() tea.Cmd {
	ch := m.stream
	return func() tea.Msg {
		step, ok := <-ch
		if !ok {
			return probeDoneMsg{}
		}
		return probeStepMsg(step)
	}
}

func (m probeModel) Update(msg tea.Msg) (probeModel, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case progress.FrameMsg:
		pm, cmd := m.bar.Update(msg)
		m.bar = pm.(progress.Model)
		return m, cmd
	case probeStepMsg:
		step := detect.ProbeStep(msg)
		m.done++
		m.current = step.Tool.Name
		m.results[step.Key] = step.Result
		return m, m.waitStep()
	case probeDoneMsg:
		final := detect.Apply(m.bundles, m.results)
		installer.SetHarnesses(detect.DetectedHarnesses())
		return m, func() tea.Msg {
			return transitionMsg{target: screenWelcome, bundles: final}
		}
	case tea.KeyMsg:
		// Allow the user to skip the probe screen with Enter / Esc.
		// Skipping early just lands on welcome with whatever results
		// have been collected; tools that didn't probe in time stay at
		// their preset defaults.
		switch msg.String() {
		case "enter", "esc":
			final := detect.Apply(m.bundles, m.results)
			installer.SetHarnesses(detect.DetectedHarnesses())
			return m, func() tea.Msg {
				return transitionMsg{target: screenWelcome, bundles: final}
			}
		}
	}
	return m, nil
}

func (m probeModel) View(width, height int) string {
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
	if m.total > 0 {
		pct = float64(m.done) / float64(m.total)
	}

	var b strings.Builder

	b.WriteString(SectionLabel(p, "Detecting installed tools",
		fmt.Sprintf("%d of %d", m.done, m.total), contentW))
	b.WriteString("\n\n")

	b.WriteString(renderStatCells(p, contentW, []statCell{
		{label: "PROBED", value: fmt.Sprintf("%02d", m.done), color: p.Primary},
		{label: "TOTAL", value: fmt.Sprintf("%02d", m.total), color: p.Text},
		{label: "PCT", value: fmt.Sprintf("%d%%", int(pct*100)), color: p.Accent},
	}))
	b.WriteString("\n\n")

	b.WriteString("  ")
	b.WriteString(m.bar.ViewAs(pct))
	b.WriteString("\n\n")

	current := lipgloss.NewStyle().Foreground(p.Muted).Italic(true).Render("starting…")
	if m.current != "" {
		bold := lipgloss.NewStyle().Foreground(p.Primary).Bold(true)
		current = fmt.Sprintf("%s checking %s", m.spinner.View(), bold.Render(m.current))
	}
	b.WriteString("  " + current + "\n")

	return Frame(p, width, height,
		"detecting",
		b.String(),
		HintLine(p,
			KeyHint(p, "⏎", "skip"),
			KeyHint(p, "^C", "quit"),
		),
		height < 22,
	)
}

func max2(a, b int) int {
	if a > b {
		return a
	}
	return b
}
