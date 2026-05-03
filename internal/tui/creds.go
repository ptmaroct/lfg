package tui

import (
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/ptmaroct/lfg/internal/preset"
)

// credsModel — credentials wizard rendered between confirm and progress.
//
// For every selected MCP tool that declares EnvVars, lfg shows one
// password-style input per unique env var (deduped across tools, so
// e.g. two tools both needing GITHUB_PERSONAL_ACCESS_TOKEN share the
// same prompt). On submit the collected map flows back into the
// selected tools' MCPCredentials field; the installer then bakes the
// values into the harness configs (via `--env KEY=VAL` for stdio,
// expanded `${VAR}` placeholders for remote URL/headers).
//
// Empty inputs are intentional: the user might want to set the var in
// their shell rc instead. The done screen reminds them of every var
// that ended up blank.
//
// Skipped entirely when no selected tool needs any env var.
type credsModel struct {
	palette  Palette
	form     *huh.Form
	keys     []string            // unique env var names, sorted
	values   map[string]*string  // bound pointers per key (huh accessor)
	bundles  []preset.Bundle     // for help text + apply step
	selected map[string]bool
	skipped  bool
}

// neededCreds returns the unique sorted set of env var names from every
// selected MCP tool. Empty result → wizard is unnecessary.
func neededCreds(bundles []preset.Bundle, selected map[string]bool) []string {
	seen := map[string]bool{}
	var out []string
	for _, b := range bundles {
		for _, t := range b.Tools {
			if t.Source != "mcp" {
				continue
			}
			if !selected[b.ID+"/"+t.Name] {
				continue
			}
			for _, ev := range t.EnvVars {
				if seen[ev] {
					continue
				}
				seen[ev] = true
				out = append(out, ev)
			}
		}
	}
	sort.Strings(out)
	return out
}

func newCreds(p Palette, bundles []preset.Bundle, selected map[string]bool) credsModel {
	keys := neededCreds(bundles, selected)
	m := credsModel{
		palette:  p,
		keys:     keys,
		values:   map[string]*string{},
		bundles:  bundles,
		selected: selected,
	}
	if len(keys) == 0 {
		m.skipped = true
		return m
	}
	fields := make([]huh.Field, 0, len(keys))
	for _, k := range keys {
		v := ""
		m.values[k] = &v
		fields = append(fields,
			huh.NewInput().
				Title(k).
				Description(credentialHelp(k)).
				EchoMode(huh.EchoModePassword).
				Placeholder("paste secret · leave blank to set later").
				Value(&v),
		)
	}
	m.form = huh.NewForm(huh.NewGroup(fields...)).
		WithTheme(HuhTheme(p)).
		WithShowHelp(false).
		WithShowErrors(false)
	return m
}

// credentialHelp returns a short hint about where to grab a key for
// the well-known env vars. Falls back to a generic message for unknown
// names so custom-preset entries still render cleanly.
func credentialHelp(name string) string {
	switch name {
	case "PERPLEXITY_API_KEY":
		return "https://www.perplexity.ai/settings/api"
	case "EXA_API_KEY":
		return "https://dashboard.exa.ai/api-keys"
	case "CONTEXT7_API_KEY":
		return "https://context7.com/dashboard (free tier works without one)"
	case "CIRCLECI_TOKEN":
		return "https://app.circleci.com/settings/user/tokens"
	case "GITHUB_PERSONAL_ACCESS_TOKEN":
		return "https://github.com/settings/tokens"
	}
	return "value will be passed via --env to the harness when registering this MCP"
}

func (m credsModel) Init() tea.Cmd {
	if m.form == nil {
		return nil
	}
	return m.form.Init()
}

func (m credsModel) Update(msg tea.Msg) (credsModel, tea.Cmd) {
	if m.skipped {
		return m, goTo(screenProgress)
	}
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "esc":
			return m, goTo(screenConfirm)
		case "ctrl+s":
			// Power-user shortcut to skip the rest of the form and proceed
			// with whatever's been entered so far.
			return m, m.applyAndAdvance()
		}
	}
	f, cmd := m.form.Update(msg)
	if ff, ok := f.(*huh.Form); ok {
		m.form = ff
	}
	if m.form.State == huh.StateCompleted {
		return m, m.applyAndAdvance()
	}
	return m, cmd
}

// applyAndAdvance writes the collected values back onto every selected
// MCP tool's MCPCredentials map, then transitions to progress with the
// mutated bundles.
func (m credsModel) applyAndAdvance() tea.Cmd {
	mutated := applyCreds(m.bundles, m.selected, m.values)
	return func() tea.Msg {
		return transitionMsg{target: screenProgress, bundles: mutated}
	}
}

// applyCreds returns a deep-enough copy of bundles where every selected
// MCP tool gets its MCPCredentials populated from the wizard's value
// map. We copy because Tool entries are passed by value through the
// installer plan; mutating in place would still work but copying keeps
// the data flow obvious in tests.
func applyCreds(bundles []preset.Bundle, selected map[string]bool, vals map[string]*string) []preset.Bundle {
	out := make([]preset.Bundle, len(bundles))
	for i, b := range bundles {
		nb := b
		nb.Tools = make([]preset.Tool, len(b.Tools))
		copy(nb.Tools, b.Tools)
		for j, t := range nb.Tools {
			if t.Source != "mcp" || !selected[b.ID+"/"+t.Name] {
				continue
			}
			creds := map[string]string{}
			for _, ev := range t.EnvVars {
				if p, ok := vals[ev]; ok && p != nil && *p != "" {
					creds[ev] = *p
				}
			}
			if len(creds) > 0 {
				nb.Tools[j].MCPCredentials = creds
			}
		}
		out[i] = nb
	}
	return out
}

func (m credsModel) View(width, height int) string {
	p := m.palette
	canvasW := CanvasW(width)
	contentW := canvasW - 4

	var b strings.Builder
	b.WriteString(SectionLabel(p, "MCP credentials", "step 2.5/2 · paste keys", contentW))
	b.WriteString("\n\n")

	if m.skipped {
		b.WriteString("  " + lipgloss.NewStyle().Foreground(p.Muted).Italic(true).
			Render("no credentials needed — proceeding to install") + "\n")
		return Frame(p, width, height,
			"step 2.5 · creds",
			b.String(),
			HintLine(p, KeyHint(p, "⏎", "continue")),
			height < 22,
		)
	}

	b.WriteString(lipgloss.NewStyle().Foreground(p.Muted).
		Render("  Paste each MCP server's API key. Inputs are masked. Leave\n  blank to skip — you can `export FOO=...` in your shell rc later."))
	b.WriteString("\n\n")
	b.WriteString(m.form.View())

	return Frame(p, width, height,
		"step 2.5 · creds",
		b.String(),
		HintLine(p,
			KeyHint(p, "⏎", "next field / submit"),
			KeyHint(p, "^S", "skip rest"),
			KeyHint(p, "ESC", "back"),
		),
		height < 22,
	)
}
