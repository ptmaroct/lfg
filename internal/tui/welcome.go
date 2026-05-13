package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ptmaroct/lfg/internal/preset"
)

// welcomeModel — first-screen entry. Custom render (not huh) so we can
// place the animated logo hero AND the numbered action list in one
// composition. The logo gradient sweeps via a tick-driven `phase`
// counter (~220 ms cadence — calm, joyful breathing pace, ~10 s full
// cycle. Fast sweeps on saturated hues felt strobe-y).
type welcomeModel struct {
	palette Palette
	cursor  int
	choices []welcomeChoice
	phase   int // animation offset for gradient sweep
}

type welcomeChoice struct {
	label  string
	desc   string
	action screen
}

// logoTickMsg fires periodically to advance the gradient sweep on the
// welcome banner. Carried by tickCmd below.
type logoTickMsg struct{}

func logoTickCmd() tea.Cmd {
	return tea.Tick(220*time.Millisecond, func(time.Time) tea.Msg {
		return logoTickMsg{}
	})
}

// remotePinsAppliedMsg fires after the async pin-fetcher lands a
// fresher PinSet from GitHub raw. Triggers a re-render so the
// freshness chip in the welcome chrome updates.
type remotePinsAppliedMsg struct{}

// remotePinsFetchCmd fires once on welcome Init. 2s timeout inside
// FetchRemotePins; on failure we silently keep the embedded pins.
func remotePinsFetchCmd() tea.Cmd {
	return func() tea.Msg {
		ps, err := preset.FetchRemotePins(context.Background(), preset.DefaultRemotePinsURL)
		if err != nil {
			return nil
		}
		preset.SetRemotePins(ps)
		return remotePinsAppliedMsg{}
	}
}

func newWelcome(p Palette) welcomeModel {
	return welcomeModel{
		palette: p,
		choices: []welcomeChoice{
			{
				label:  "INSTALL RECOMMENDED SETUP",
				desc:   "Pick bundles, install tools. Opinionated but customizable.",
				action: screenTree,
			},
			{
				label:  "LOAD CONFIG FILE",
				desc:   "Open a dialog to paste a preset path or URL.",
				action: screenConfigInput,
			},
			{
				label:  "EXPORT THIS MACHINE AS PRESET",
				desc:   "Save the current bundle/tool set to a TOML file you can re-load elsewhere.",
				action: screenExport,
			},
			{
				label:  "BACKUP THIS MACHINE",
				desc:   "Snapshot dotfiles, package list, configs into a single file.",
				action: screenBackupPrompt,
			},
			{
				label:  "QUIT",
				desc:   "Exit lfg.",
				action: screenQuit,
			},
		},
	}
}

func (m welcomeModel) Init() tea.Cmd {
	return tea.Batch(logoTickCmd(), remotePinsFetchCmd())
}

func (m welcomeModel) Update(msg tea.Msg) (welcomeModel, tea.Cmd) {
	switch msg := msg.(type) {
	case logoTickMsg:
		m.phase = (m.phase + 1) % 1000 // any large number; mod inside renderer
		return m, logoTickCmd()
	case remotePinsAppliedMsg:
		// Pin freshness chip now reflects the freshly-fetched set; no
		// state change needed since CurrentPins() is consulted at render.
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			return m, goTo(m.choices[m.cursor].action)
		case "esc", "backspace", "delete":
			return m, goTo(screenQuitConfirm)
		case "1", "2", "3", "4", "5":
			idx := int(msg.String()[0] - '1')
			if idx < len(m.choices) {
				m.cursor = idx
				return m, goTo(m.choices[idx].action)
			}
		}
	}
	return m, nil
}

func (m welcomeModel) View(width, height int) string {
	p := m.palette

	canvasW := CanvasW(width)
	contentW := canvasW - 4
	compact := height < 22

	var b strings.Builder

	// Hero block — animated gradient logo on big terminals, single-line on small.
	hero := RenderTitle(p, "lfg", compact, m.phase)
	b.WriteString(lipgloss.PlaceHorizontal(contentW, lipgloss.Center, hero))
	b.WriteString("\n\n")
	tagline := lipgloss.NewStyle().Foreground(p.Muted).Italic(true).
		Render("a new dev machine, in less time than this hint takes to read")
	b.WriteString(lipgloss.PlaceHorizontal(contentW, lipgloss.Center, tagline))
	b.WriteString("\n")
	b.WriteString(lipgloss.PlaceHorizontal(contentW, lipgloss.Center, renderPinFreshness(p)))
	b.WriteString("\n\n")

	// Actions section
	b.WriteString(SectionLabel(p, "What now?", "fresh machine detected", contentW))
	b.WriteString("\n\n")

	for i, c := range m.choices {
		b.WriteString(m.renderChoice(i, c, contentW))
		b.WriteString("\n")
	}

	return Frame(p, width, height,
		"welcome",
		b.String(),
		HintLine(p,
			KeyHint(p, "↑↓", "nav"),
			KeyHint(p, "1-5", "jump"),
			KeyHint(p, "⏎", "select"),
			KeyHint(p, "^T", "theme"),
			KeyHint(p, "Q", "quit"),
		),
		compact,
	)
}

func (m welcomeModel) renderChoice(i int, c welcomeChoice, contentW int) string {
	p := m.palette
	num := lipgloss.NewStyle().Foreground(p.Muted).Render(strings.Repeat("0", 1) + sprint1(i+1))
	gutter := "  "
	titleStyle := lipgloss.NewStyle().Foreground(p.Text).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(p.Muted).Italic(true)

	if i == m.cursor {
		// Cursor stays Primary for accent identity. Number flips to Text
		// bold (high contrast) instead of Primary — colored small digits
		// on dark bgs read dim because mid-saturation hues lose
		// luminance vs. plain bright text. Title stays Primary so the
		// row's selection signal is unmistakable.
		gutter = lipgloss.NewStyle().Foreground(p.Primary).Bold(true).Render("▸ ")
		num = lipgloss.NewStyle().Foreground(p.Text).Bold(true).Render("0" + sprint1(i+1))
		titleStyle = titleStyle.Foreground(p.Primary)
	}

	line1 := gutter + num + "  " + titleStyle.Render(c.label)
	// Description sits exactly under the label: gutter(2) + num(2) +
	// gap(2) = 6 spaces before label starts, so 6 before desc too.
	line2 := strings.Repeat(" ", 6) + descStyle.Render(c.desc)
	_ = contentW
	return line1 + "\n" + line2
}

// renderPinFreshness draws the "pins: 2026-05-13 (3 d)" chip under
// the welcome tagline. Colour shifts as the embedded pin set ages,
// nudging long-running users to `lfg update` once they cross the
// 30-day mark. Hidden entirely when no pin set is loaded (e.g. fresh
// fork, custom config) so we don't shout "unknown" at first-run.
func renderPinFreshness(p Palette) string {
	ps := preset.CurrentPins()
	if ps.BumpedAt.IsZero() {
		return ""
	}
	ageDays, bucket := preset.PinFreshness()
	colour := p.Success
	suffix := ""
	switch bucket {
	case "stale":
		colour = p.Accent
		suffix = " — consider `lfg update`"
	case "very-stale":
		colour = p.Primary
		suffix = " — run `lfg update`"
	}
	label := fmt.Sprintf("pins: %s (%d d)%s", ps.BumpedAt.Format("2006-01-02"), ageDays, suffix)
	return lipgloss.NewStyle().Foreground(colour).Render(label)
}

// sprint1 — tiny int→string for 1..9. Avoids importing strconv.
func sprint1(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return ""
}
