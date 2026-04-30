package tui

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/stopwatch"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/anuj/lfg/internal/installer"
	"github.com/anuj/lfg/internal/preset"
)

// ProgressRunner is the function signature progressModel uses to drive
// installs. The CLI passes installer.Run; tests + the snap helper pass
// a mock that emits canned lines without touching the host system.
type ProgressRunner func(ctx context.Context, plan []installer.Step, out chan<- installer.Line) []installer.FailedStep

// progressModel — instrument-panel install screen.
//
// Stat row top, gradient progress bar, current-action line, log tail.
// Drives a real installer.Run via the injected ProgressRunner; lines
// stream into the log tail and `begin`/`end` markers advance the bar.
type progressModel struct {
	palette   Palette
	plan      []installer.Step
	queue     []preset.Tool // tools (bootstraps excluded) — for header counts
	currentT  string        // bundle/tool key of in-flight step
	index     int           // step index (0..len(plan))
	logs      []string
	failed    []installer.FailedStep
	spinner   spinner.Model
	bar       progress.Model
	stopwatch stopwatch.Model
	done      bool

	runner ProgressRunner
	lineCh chan installer.Line
	doneCh chan []installer.FailedStep
	cancel context.CancelFunc
}

// installLineMsg wraps a streamed installer.Line for the bubbletea loop.
type installLineMsg installer.Line

// installFinishedMsg is delivered exactly once when the installer
// goroutine returns, carrying the final failure list.
type installFinishedMsg struct {
	failed []installer.FailedStep
}

// installStepMsg is retained as a no-op type so older snapshot tests
// importing the symbol continue to compile.
type installStepMsg struct{ idx int }

// mockProgressRunner emits a synthetic begin/end pair per step with
// short pauses so screenshots / snapshot tests get realistic-looking
// state without running any subprocess.
func mockProgressRunner(ctx context.Context, plan []installer.Step, out chan<- installer.Line) []installer.FailedStep {
	for _, s := range plan {
		key := s.Backend
		if !s.Bootstrap {
			key = s.Bundle + "/" + s.Tool.Name
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Duration(300+rand.Intn(600)) * time.Millisecond):
		}
		out <- installer.Line{Tool: key, Stream: "begin"}
		out <- installer.Line{Tool: key, Stream: "stdout", Text: "(mock) installing " + key}
		out <- installer.Line{Tool: key, Stream: "end"}
	}
	return nil
}

func newProgress(p Palette, bundles []preset.Bundle, selected map[string]bool) progressModel {
	return newProgressWithRunner(p, bundles, selected, mockProgressRunner)
}

// newProgressWithRunner builds a progress screen wired to the given
// runner. CLI startup uses installer.Run; tests + snap helper use the
// built-in mock.
func newProgressWithRunner(p Palette, bundles []preset.Bundle, selected map[string]bool, runner ProgressRunner) progressModel {
	var queue []preset.Tool
	for _, b := range bundles {
		for _, t := range b.Tools {
			if selected[b.ID+"/"+t.Name] {
				queue = append(queue, t)
			}
		}
	}
	plan := installer.Plan(bundles, selected)

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
		plan:      plan,
		queue:     queue,
		spinner:   sp,
		bar:       bar,
		stopwatch: sw,
		runner:    runner,
		lineCh:    make(chan installer.Line, 64),
		doneCh:    make(chan []installer.FailedStep, 1),
	}
}

// Init kicks off the installer goroutine and arms the line-pump.
func (m progressModel) Init() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	go func() {
		failed := m.runner(ctx, m.plan, m.lineCh)
		m.doneCh <- failed
		close(m.lineCh)
	}()
	return tea.Batch(m.spinner.Tick, m.stopwatch.Start(), m.waitLine())
}

// waitLine returns a tea.Cmd that pulls one Line off the channel. The
// command re-arms itself for each line so the loop continues until the
// channel closes.
func (m progressModel) waitLine() tea.Cmd {
	lineCh := m.lineCh
	doneCh := m.doneCh
	return func() tea.Msg {
		l, ok := <-lineCh
		if !ok {
			// Channel closed → installer goroutine has finished. Pick up
			// the failure summary and emit a finished msg.
			failed := <-doneCh
			return installFinishedMsg{failed: failed}
		}
		return installLineMsg(l)
	}
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
	case installLineMsg:
		l := installer.Line(msg)
		switch l.Stream {
		case "begin":
			m.currentT = l.Tool
		case "end":
			m.index++
			status := lipgloss.NewStyle().Foreground(m.palette.Success).Render("●")
			if l.Text != "" {
				status = lipgloss.NewStyle().Foreground(m.palette.Warn).Render("✗")
			}
			short := l.Tool
			m.logs = append(m.logs, fmt.Sprintf("%s  %-26s  done", status, short))
		case "stdout", "stderr":
			if l.Text != "" {
				dot := lipgloss.NewStyle().Foreground(m.palette.Subtle).Render("·")
				m.logs = append(m.logs, fmt.Sprintf("%s  %s", dot, truncate(l.Text, 60)))
			}
		}
		if len(m.logs) > 8 {
			m.logs = m.logs[len(m.logs)-8:]
		}
		return m, m.waitLine()
	case installFinishedMsg:
		m.done = true
		m.failed = msg.failed
		return m, tea.Batch(
			m.stopwatch.Stop(),
			tea.Tick(800*time.Millisecond, func(time.Time) tea.Msg {
				return transitionMsg{target: screenDone}
			}),
		)
	case progress.FrameMsg:
		pm, cmd := m.bar.Update(msg)
		m.bar = pm.(progress.Model)
		return m, cmd
	case tea.KeyMsg:
		if msg.String() == "esc" {
			if m.cancel != nil {
				m.cancel()
			}
			return m, goTo(screenConfirm)
		}
	}
	return m, nil
}

// truncate keeps the log tail from blowing past one terminal line.
func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
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
	if len(m.plan) > 0 {
		pct = float64(m.index) / float64(len(m.plan))
	}

	var b strings.Builder

	// Header strip already says "INSTALLING" — body section label uses
	// "STATUS" so we don't repeat the same word twice. Right-side suffix
	// carries the count (was a separate "7 OF 7" line before).
	statusLabel := "Status"
	statusSuffix := fmt.Sprintf("%d of %d", m.index, len(m.plan))
	if m.done {
		statusLabel = "Complete"
		ok := len(m.plan) - len(m.failed)
		if len(m.failed) > 0 {
			statusSuffix = fmt.Sprintf("%d ok · %d failed", ok, len(m.failed))
		} else {
			statusSuffix = fmt.Sprintf("%d installed", ok)
		}
	}
	b.WriteString(SectionLabel(p, statusLabel, statusSuffix, contentW))
	b.WriteString("\n\n")

	// Boxed stat readouts (consistent with confirm screen).
	failColor := p.Success
	if len(m.failed) > 0 {
		failColor = p.Warn
	}
	b.WriteString(renderStatCells(p, contentW, []statCell{
		{label: "DONE", value: fmt.Sprintf("%02d", m.index), color: p.Primary},
		{label: "TOTAL", value: fmt.Sprintf("%02d", len(m.plan)), color: p.Text},
		{label: "ELAPSED", value: m.stopwatch.View(), color: p.Accent},
		{label: "PCT", value: fmt.Sprintf("%d%%", int(pct*100)), color: failColor},
	}))
	b.WriteString("\n\n")

	// Bar (full width).
	b.WriteString("  ")
	b.WriteString(m.bar.ViewAs(pct))
	b.WriteString("\n\n")

	// Current action.
	current := lipgloss.NewStyle().Foreground(p.Muted).Render("— done —")
	if !m.done && m.currentT != "" {
		bold := lipgloss.NewStyle().Foreground(p.Primary).Bold(true)
		current = fmt.Sprintf("%s %s", m.spinner.View(), bold.Render(m.currentT))
	} else if m.done {
		if len(m.failed) > 0 {
			current = lipgloss.NewStyle().Foreground(p.Warn).Bold(true).
				Render(fmt.Sprintf("● done with %d failure(s)", len(m.failed)))
		} else {
			current = lipgloss.NewStyle().Foreground(p.Success).Bold(true).Render("● all installed")
		}
	}
	b.WriteString("  " + current + "\n\n")

	// Log tail — section label header (matches the rest of the screen),
	// hairlines top + bottom for clear bounding.
	logLabel := lipgloss.NewStyle().Foreground(p.Muted).Render("LOG")
	logCount := lipgloss.NewStyle().Foreground(p.Subtle).
		Render(fmt.Sprintf("· last %d", len(m.logs)))
	b.WriteString("  " + logLabel + "  " + logCount + "\n")
	b.WriteString("  " + Hairline(p, contentW-2) + "\n")
	if len(m.logs) == 0 {
		b.WriteString("  " + lipgloss.NewStyle().Foreground(p.Muted).Italic(true).
			Render("(no entries yet)") + "\n")
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
