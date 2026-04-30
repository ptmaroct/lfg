package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/anuj/lfg/internal/preset"
)

// bundlePickerModel uses huh.MultiSelect to stack bundles.
type bundlePickerModel struct {
	palette  Palette
	form     *huh.Form
	bundles  []preset.Bundle
	selected []string
}

func newBundlePicker(p Palette, bundles []preset.Bundle, preselected map[string]bool) bundlePickerModel {
	m := bundlePickerModel{palette: p, bundles: bundles}

	opts := make([]huh.Option[string], 0, len(bundles))
	for _, b := range bundles {
		// Compact label: name + tool count. Description rendered separately
		// in the form (huh shows it under the field title), avoiding the
		// long-line wrap problem we saw when stuffing description in label.
		label := fmt.Sprintf("%-16s %d tools", b.Name, len(b.Tools))
		opts = append(opts, huh.NewOption(label, b.ID).Selected(preselected[b.ID]))
	}

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Pick bundles to stack").
				Description("Each bundle brings a curated set of tools. default is pre-selected.").
				Options(opts...).
				Value(&m.selected).
				Height(10).
				Validate(func(v []string) error {
					if len(v) == 0 {
						return fmt.Errorf("pick at least one bundle")
					}
					return nil
				}),
		),
	).
		WithTheme(HuhTheme(p)).
		WithShowHelp(false).
		WithShowErrors(true)

	return m
}

func (m bundlePickerModel) Init() tea.Cmd { return m.form.Init() }

func (m bundlePickerModel) Update(msg tea.Msg) (bundlePickerModel, tea.Cmd) {
	f, cmd := m.form.Update(msg)
	if ff, ok := f.(*huh.Form); ok {
		m.form = ff
	}
	if m.form.State == huh.StateCompleted {
		ids := map[string]bool{}
		for _, id := range m.selected {
			ids[id] = true
		}
		// Pre-populate tool selections: pre-check everything not yet installed.
		tools := map[string]bool{}
		for _, b := range m.bundles {
			if !ids[b.ID] {
				continue
			}
			for _, t := range b.Tools {
				tools[b.ID+"/"+t.Name] = !t.Installed
			}
		}
		return m, func() tea.Msg {
			return transitionMsg{
				target:            screenTools,
				selectedBundleIDs: ids,
				selectedTools:     tools,
			}
		}
	}
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "esc" {
		return m, goTo(screenWelcome)
	}
	return m, cmd
}

func (m bundlePickerModel) View(width, height int) string {
	return Frame(m.palette, width, height,
		"step 1 of 3  ·  bundles",
		m.form.View(),
		HintLine(m.palette,
			KeyHint(m.palette, "↑/↓", "move"),
			KeyHint(m.palette, "x", "toggle"),
			KeyHint(m.palette, "enter", "continue"),
			KeyHint(m.palette, "esc", "back"),
		),
		height < 22,
	)
}
