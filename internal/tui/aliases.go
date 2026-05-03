package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/ptmaroct/lfg/internal/preset"
)

// aliasesModel — shell-alias picker rendered after install (between
// progress and done). Lowest priority in the setup flow: aliases are
// quality-of-life, never blocking. Skipped entirely when there are no
// aliases to offer (empty config or DefaultAliases returned nothing).
//
// Single huh.MultiSelect with options grouped visually by sorting on
// AliasGroup; conflicts (alias name already defined in a user rc) get
// rendered as a `⚠` suffix on the option label so the user sees the
// risk before they hit submit.
type aliasesModel struct {
	palette   Palette
	form      *huh.Form
	groups    []preset.AliasGroup
	flat      []preset.Alias
	conflicts map[string]string
	keys      []string // option key per visible alias, indexed parallel to flat
	selected  []string // bound by huh.MultiSelect
	skipped   bool
}

// flattenGroups returns the aliases in canonical group order so the
// picker renders gits together, claudes together, shell helpers last.
func flattenGroups(groups []preset.AliasGroup) []preset.Alias {
	var out []preset.Alias
	for _, g := range groups {
		out = append(out, g.Aliases...)
	}
	return out
}

func newAliases(p Palette, groups []preset.AliasGroup, preselected map[string]bool, conflicts map[string]string) aliasesModel {
	flat := flattenGroups(groups)
	m := aliasesModel{
		palette:   p,
		groups:    groups,
		flat:      flat,
		conflicts: conflicts,
		keys:      make([]string, 0, len(flat)),
	}
	if len(flat) == 0 {
		m.skipped = true
		return m
	}

	opts := make([]huh.Option[string], 0, len(flat))
	for _, a := range flat {
		key := a.Group + "/" + a.Name
		m.keys = append(m.keys, key)
		opts = append(opts, huh.NewOption(aliasOptionLabel(a, conflicts), key).
			Selected(aliasIsSelected(a, key, preselected)))
	}

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Pick shell aliases").
				Description("Selected aliases get added to a fenced lfg-managed block in your shell rc(s).\nYou can re-run lfg anytime to change selection.").
				Options(opts...).
				Value(&m.selected).
				Height(min(14, len(opts)+2)),
		),
	).
		WithTheme(HuhTheme(p)).
		WithShowHelp(false).
		WithShowErrors(false)
	return m
}

// aliasIsSelected returns true if either the user's prior selection
// includes the key or the alias is a default and there's no prior
// selection map (empty map = first visit, honor defaults).
func aliasIsSelected(a preset.Alias, key string, pre map[string]bool) bool {
	if len(pre) > 0 {
		return pre[key]
	}
	return a.Default
}

// aliasOptionLabel formats one option line for the picker.
//
// Layout: `<NAME>  <description> · <command>`. When a conflict is
// detected (name already defined in a user rc outside the lfg block)
// we suffix `⚠ in <rc>:<line>` so the user knows what they'd shadow.
func aliasOptionLabel(a preset.Alias, conflicts map[string]string) string {
	desc := a.Description
	if desc == "" {
		desc = a.Command
	}
	label := fmt.Sprintf("%-7s %s", a.Name, desc)
	if conflicts != nil {
		if where, ok := conflicts[a.Name]; ok {
			label += fmt.Sprintf("  ⚠ in %s", where)
		}
	}
	return label
}

func (m aliasesModel) Init() tea.Cmd {
	if m.skipped || m.form == nil {
		return nil
	}
	return m.form.Init()
}

func (m aliasesModel) Update(msg tea.Msg) (aliasesModel, tea.Cmd) {
	if m.skipped {
		if k, ok := msg.(tea.KeyMsg); ok {
			switch k.String() {
			case "enter", "esc", " ":
				return m, goTo(screenDone)
			}
		}
		return m, nil
	}
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "esc":
			// Skip — proceed without writing aliases.
			return m, m.advance(nil)
		case "ctrl+s":
			// Power-user shortcut: submit current selection.
			return m, m.advance(m.selectedSet())
		}
	}
	f, cmd := m.form.Update(msg)
	if ff, ok := f.(*huh.Form); ok {
		m.form = ff
	}
	if m.form.State == huh.StateCompleted {
		return m, m.advance(m.selectedSet())
	}
	return m, cmd
}

// selectedSet converts huh's []string slice into the keyed map the
// root model carries.
func (m aliasesModel) selectedSet() map[string]bool {
	out := map[string]bool{}
	for _, k := range m.selected {
		out[k] = true
	}
	return out
}

// advance returns a cmd that hops to screenDone with the alias
// selection threaded through. nil → user skipped, no aliases written.
func (m aliasesModel) advance(set map[string]bool) tea.Cmd {
	return func() tea.Msg {
		return transitionMsg{
			target:          screenDone,
			selectedAliases: set,
		}
	}
}

func (m aliasesModel) View(width, height int) string {
	p := m.palette
	canvasW := CanvasW(width)
	contentW := canvasW - 4

	var b strings.Builder
	b.WriteString(SectionLabel(p, "Shell aliases", "optional · written to your rc", contentW))
	b.WriteString("\n\n")

	if m.skipped {
		b.WriteString("  " + lipgloss.NewStyle().Foreground(p.Muted).Italic(true).
			Render("no aliases offered — proceeding") + "\n")
		return Frame(p, width, height,
			"step 4 · aliases",
			b.String(),
			HintLine(p, KeyHint(p, "⏎", "continue")),
			height < 22,
		)
	}

	intro := lipgloss.NewStyle().Foreground(p.Muted).
		Render("  Quality-of-life shortcuts. We write a fenced block to every detected\n  shell rc; deselecting all leaves your rc unchanged.")
	b.WriteString(intro + "\n\n")

	if len(m.conflicts) > 0 {
		warn := lipgloss.NewStyle().Foreground(p.Warn).Bold(true).Render("  ⚠ ")
		warnText := lipgloss.NewStyle().Foreground(p.Muted).Render(
			"some names already defined in your rc — selecting will shadow the existing alias")
		b.WriteString(warn + warnText + "\n\n")
	}

	b.WriteString(m.form.View())

	return Frame(p, width, height,
		"step 4 · aliases",
		b.String(),
		HintLine(p,
			KeyHint(p, "x", "toggle"),
			KeyHint(p, "⏎", "continue"),
			KeyHint(p, "^S", "submit now"),
			KeyHint(p, "ESC", "skip"),
		),
		height < 22,
	)
}

// resolveSelectedAliases walks the alias catalog and returns the
// concrete []preset.Alias the user picked. Order preserved per the
// catalog's group ordering (so the rc block is stable across runs).
func resolveSelectedAliases(groups []preset.AliasGroup, selected map[string]bool) []preset.Alias {
	if len(selected) == 0 {
		return nil
	}
	var out []preset.Alias
	for _, g := range groups {
		for _, a := range g.Aliases {
			if selected[a.Group+"/"+a.Name] {
				out = append(out, a)
			}
		}
	}
	return out
}
