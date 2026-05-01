package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/ptmaroct/lfg/internal/preset"
)

// toolPickerModel shows one huh.MultiSelect per bundle as separate groups.
// huh 1.0 renders one group at a time — user tabs between groups with tab/shift-tab.
type toolPickerModel struct {
	palette   Palette
	form      *huh.Form
	selected  map[string][]string // bundleID -> []toolName
	bundleIDs []string             // stable order
	bundles   map[string]preset.Bundle
}

func newToolPicker(p Palette, bundles []preset.Bundle, bundleIDs map[string]bool, preselected map[string]bool) toolPickerModel {
	m := toolPickerModel{
		palette:  p,
		selected: map[string][]string{},
		bundles:  map[string]preset.Bundle{},
	}

	// Preserve bundle order from preset.All().
	var groups []*huh.Group
	for _, b := range bundles {
		if !bundleIDs[b.ID] {
			continue
		}
		m.bundleIDs = append(m.bundleIDs, b.ID)
		m.bundles[b.ID] = b

		opts := make([]huh.Option[string], 0, len(b.Tools))
		for _, t := range b.Tools {
			label := formatToolLabel(t, p)
			opt := huh.NewOption(label, t.Name).
				Selected(preselected[b.ID+"/"+t.Name])
			opts = append(opts, opt)
		}

		// Per-bundle selection slice (huh writes into each slice).
		sel := []string{}
		// Pre-fill with preselected for this bundle
		for _, t := range b.Tools {
			if preselected[b.ID+"/"+t.Name] {
				sel = append(sel, t.Name)
			}
		}
		m.selected[b.ID] = sel
		selRef := m.selected[b.ID]
		_ = selRef

		// NOTE: huh.NewMultiSelect writes into a *[]string, so we need a
		// stable pointer. Index into the map directly via closure-safe var.
		bundleID := b.ID
		slicePtr := func() *[]string {
			s := m.selected[bundleID]
			return &s
		}()

		ms := huh.NewMultiSelect[string]().
			Title(fmt.Sprintf("Tools · %s", b.Name)).
			Description(b.Description).
			Options(opts...).
			Value(slicePtr).
			Filterable(true).
			Height(12)

		groups = append(groups, huh.NewGroup(ms))

		// Tie the pointer back into the map so its updates are visible on submit.
		// (huh mutates the slice in place; we need to make sure the map sees it.)
		m.selected[bundleID] = *slicePtr
	}

	m.form = huh.NewForm(groups...).
		WithTheme(HuhTheme(p)).
		WithShowHelp(false).
		WithShowErrors(true)

	return m
}

// formatToolLabel — `name  ✓ version  · brew` or `name  (not installed)  · brew`.
func formatToolLabel(t preset.Tool, p Palette) string {
	name := lipgloss.NewStyle().Foreground(p.Text).Render(t.Name)
	name = padRight(name, 24, t.Name)

	var status string
	if t.Installed {
		status = lipgloss.NewStyle().Foreground(p.Success).Render("✓ " + t.Version)
	} else {
		status = lipgloss.NewStyle().Foreground(p.Muted).Italic(true).Render("not installed")
	}
	status = padRight(status, 18, stripANSI(status))

	src := lipgloss.NewStyle().Foreground(p.Muted).Render("· " + t.Source)
	return name + " " + status + " " + src
}

// padRight pads the rendered string to visualLen using spaces, based on
// the already-rendered content's width (measured from the stripped plain form).
func padRight(rendered string, target int, plain string) string {
	w := lipgloss.Width(plain)
	if w >= target {
		return rendered
	}
	pad := target - w
	return rendered + spaces(pad)
}

func spaces(n int) string {
	if n <= 0 {
		return ""
	}
	b := make([]byte, n)
	for i := range b {
		b[i] = ' '
	}
	return string(b)
}

// stripANSI is a very small helper — huh labels don't need full ANSI parsing,
// we only pass raw text here. Kept for padRight's API.
func stripANSI(s string) string { return s }

func (m toolPickerModel) Init() tea.Cmd { return m.form.Init() }

func (m toolPickerModel) Update(msg tea.Msg) (toolPickerModel, tea.Cmd) {
	f, cmd := m.form.Update(msg)
	if ff, ok := f.(*huh.Form); ok {
		m.form = ff
	}

	if m.form.State == huh.StateCompleted {
		tools := map[string]bool{}
		for _, bid := range m.bundleIDs {
			for _, tname := range m.selected[bid] {
				tools[bid+"/"+tname] = true
			}
		}
		return m, goToWithTools(screenConfirm, tools)
	}
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "esc" {
		return m, goTo(screenBundles)
	}
	return m, cmd
}

func (m toolPickerModel) View(width, height int) string {
	return Frame(m.palette, width, height,
		"step 2 of 3  ·  tools",
		m.form.View(),
		HintLine(m.palette,
			KeyHint(m.palette, "↑/↓", "move"),
			KeyHint(m.palette, "x", "toggle"),
			KeyHint(m.palette, "/", "filter"),
			KeyHint(m.palette, "tab", "next bundle"),
			KeyHint(m.palette, "enter", "submit"),
			KeyHint(m.palette, "esc", "back"),
		),
		height < 22,
	)
}
