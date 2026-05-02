package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ptmaroct/lfg/internal/preset"
)

// treePickerModel — single-screen tree view replacing bundle + tool pickers.
//
// Aesthetic: tabular bulletin. ID column · NAME column · STATUS dot · COUNT.
// Hairline rules between header and rows. Cursor is a colored ▸ in the
// gutter (no row highlight); selected items light up via Primary text.
//
// Selection works at both levels:
//   - toggle on a tool   → flip just that tool
//   - toggle on a bundle → flip all tools in the bundle (all-on or all-off)
//
// Bundle status dot: ● all on · ○ none · ◐ partial.
type treePickerModel struct {
	palette  Palette
	bundles  []preset.Bundle
	rows     []treeRow
	cursor   int
	expanded map[string]bool
	selected map[string]bool
	offset   int
	pageSize int
}

type treeRow struct {
	kind     string // "bundle" or "tool"
	bundleID string
	toolName string
	depth    int
	idx      int // index within parent bundle (1-based for tools)
}

func newTreePicker(p Palette, bundles []preset.Bundle, initialBundleIDs map[string]bool, initialTools map[string]bool) treePickerModel {
	m := treePickerModel{
		palette:  p,
		bundles:  bundles,
		expanded: map[string]bool{},
		selected: map[string]bool{},
		pageSize: 16,
	}
	for k, v := range initialTools {
		m.selected[k] = v
	}
	if len(initialTools) == 0 {
		for _, b := range bundles {
			if !initialBundleIDs[b.ID] {
				continue
			}
			for _, t := range b.Tools {
				m.selected[b.ID+"/"+t.Name] = !t.Installed
			}
		}
	}
	// All bundles start collapsed by default. Cleaner first impression.
	m.rebuildRows()
	return m
}

func (m *treePickerModel) rebuildRows() {
	m.rows = m.rows[:0]
	for _, b := range m.bundles {
		m.rows = append(m.rows, treeRow{kind: "bundle", bundleID: b.ID})
		if m.expanded[b.ID] {
			for i, t := range b.Tools {
				m.rows = append(m.rows, treeRow{
					kind: "tool", bundleID: b.ID, toolName: t.Name,
					depth: 1, idx: i + 1,
				})
			}
		}
	}
	if m.cursor >= len(m.rows) {
		m.cursor = len(m.rows) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m treePickerModel) bundleSelectionState(b preset.Bundle) string {
	on, off := 0, 0
	for _, t := range b.Tools {
		if m.selected[b.ID+"/"+t.Name] {
			on++
		} else {
			off++
		}
	}
	switch {
	case on == 0:
		return "none"
	case off == 0:
		return "all"
	default:
		return "partial"
	}
}

func (m treePickerModel) bundleSelectedCount(b preset.Bundle) int {
	n := 0
	for _, t := range b.Tools {
		if m.selected[b.ID+"/"+t.Name] {
			n++
		}
	}
	return n
}

func (m *treePickerModel) toggleAtCursor() {
	row := m.rows[m.cursor]
	if row.kind == "tool" {
		key := row.bundleID + "/" + row.toolName
		m.selected[key] = !m.selected[key]
		return
	}
	bundle := m.findBundle(row.bundleID)
	target := m.bundleSelectionState(bundle) != "all"
	for _, t := range bundle.Tools {
		m.selected[bundle.ID+"/"+t.Name] = target
	}
}

func (m treePickerModel) findBundle(id string) preset.Bundle {
	for _, b := range m.bundles {
		if b.ID == id {
			return b
		}
	}
	return preset.Bundle{}
}

func (m treePickerModel) findTool(bundleID, toolName string) preset.Tool {
	for _, b := range m.bundles {
		if b.ID != bundleID {
			continue
		}
		for _, t := range b.Tools {
			if t.Name == toolName {
				return t
			}
		}
	}
	return preset.Tool{}
}

func (m *treePickerModel) expandAtCursor() {
	row := m.rows[m.cursor]
	if row.kind == "bundle" {
		m.expanded[row.bundleID] = true
		m.rebuildRows()
	}
}

func (m *treePickerModel) collapseAtCursor() {
	row := m.rows[m.cursor]
	switch row.kind {
	case "bundle":
		if m.expanded[row.bundleID] {
			m.expanded[row.bundleID] = false
			m.rebuildRows()
		}
	case "tool":
		m.expanded[row.bundleID] = false
		m.rebuildRows()
		for i, r := range m.rows {
			if r.kind == "bundle" && r.bundleID == row.bundleID {
				m.cursor = i
				break
			}
		}
	}
}

func (m treePickerModel) Init() tea.Cmd { return nil }

func (m treePickerModel) Update(msg tea.Msg) (treePickerModel, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.rows)-1 {
				m.cursor++
			}
		case "right", "l":
			m.expandAtCursor()
		case "left", "h":
			m.collapseAtCursor()
		case " ", "x":
			m.toggleAtCursor()
		case "a":
			anyOn := false
			for _, b := range m.bundles {
				if m.bundleSelectionState(b) != "none" {
					anyOn = true
					break
				}
			}
			target := !anyOn
			for _, b := range m.bundles {
				for _, t := range b.Tools {
					m.selected[b.ID+"/"+t.Name] = target
				}
			}
		case "enter":
			row := m.rows[m.cursor]
			if row.kind == "bundle" && !m.expanded[row.bundleID] {
				m.expandAtCursor()
				return m, nil
			}
			if m.totalSelected() == 0 {
				return m, nil
			}
			tools := map[string]bool{}
			for k, v := range m.selected {
				if v {
					tools[k] = true
				}
			}
			return m, goToWithTools(screenConfirm, tools)
		case "esc":
			return m, goTo(screenWelcome)
		}
		if m.cursor < m.offset {
			m.offset = m.cursor
		}
		if m.cursor >= m.offset+m.pageSize {
			m.offset = m.cursor - m.pageSize + 1
		}
	}
	return m, nil
}

func (m treePickerModel) totalSelected() int {
	n := 0
	for _, v := range m.selected {
		if v {
			n++
		}
	}
	return n
}

func (m treePickerModel) View(width, height int) string {
	p := m.palette
	canvasW := width - 4
	if canvasW > 100 {
		canvasW = 100
	}
	if canvasW < 56 {
		canvasW = 56
	}
	contentW := canvasW - 4

	var b strings.Builder

	// Section label
	b.WriteString(SectionLabel(p, "Pick what to install", fmt.Sprintf("%d selected", m.totalSelected()), contentW))
	b.WriteString("\n\n")

	// Column header
	b.WriteString(m.renderColumnHeader(contentW))
	b.WriteByte('\n')
	b.WriteString("  " + Hairline(p, contentW-2))
	b.WriteByte('\n')

	end := min2(m.offset+m.pageSize, len(m.rows))
	for i := m.offset; i < end; i++ {
		b.WriteString(m.renderRow(i))
		b.WriteByte('\n')
	}

	b.WriteString("  " + Hairline(p, contentW-2))
	b.WriteByte('\n')

	return Frame(p, width, height,
		"step 1/2 · pick tools",
		b.String(),
		HintLine(p,
			KeyHint(p, "↑↓", "nav"),
			KeyHint(p, "→←", "tree"),
			KeyHint(p, "SP", "toggle"),
			KeyHint(p, "A", "all"),
			KeyHint(p, "⏎", "next"),
			KeyHint(p, "⎋", "back"),
		),
		height < 22,
	)
}

func (m treePickerModel) renderColumnHeader(contentW int) string {
	p := m.palette
	colStyle := lipgloss.NewStyle().Foreground(p.Muted)
	// Layout: cursor(2) + checkbox(4) + caret(2) + name(28) + status(rest)
	return fmt.Sprintf("    %s %s %s %s",
		colStyle.Render("    "),
		colStyle.Render("  "),
		colStyle.Render(padRightPlain("BUNDLE / TOOL", 28)),
		colStyle.Render("STATUS"),
	)
}

// checkbox returns a clear, dumb-obvious selection glyph.
//   [ ] = none, [~] = partial (bundles), [✓] = all/selected
func checkbox(p Palette, state string) string {
	switch state {
	case "all":
		return lipgloss.NewStyle().Foreground(p.Primary).Bold(true).Render("[✓]")
	case "partial":
		return lipgloss.NewStyle().Foreground(p.Warn).Bold(true).Render("[~]")
	default:
		return lipgloss.NewStyle().Foreground(p.Subtle).Render("[ ]")
	}
}

func padRightPlain(s string, n int) string {
	w := lipgloss.Width(s)
	if w >= n {
		return s
	}
	return s + strings.Repeat(" ", n-w)
}

func (m treePickerModel) renderRow(i int) string {
	p := m.palette
	row := m.rows[i]

	gutter := "  "
	if i == m.cursor {
		gutter = lipgloss.NewStyle().Foreground(p.Primary).Bold(true).Render("▸ ")
	}

	if row.kind == "bundle" {
		return gutter + m.renderBundleRow(row)
	}
	return gutter + m.renderToolRow(row)
}

// renderBundleRow:  [✓] ▼ DEFAULT                      22 tools · 22 selected
func (m treePickerModel) renderBundleRow(row treeRow) string {
	p := m.palette
	bundle := m.findBundle(row.bundleID)
	state := m.bundleSelectionState(bundle)

	box := checkbox(p, state)
	caret := "▶"
	if m.expanded[bundle.ID] {
		caret = "▼"
	}
	caretStyled := lipgloss.NewStyle().Foreground(p.Primary).Render(caret)

	name := strings.ToUpper(bundle.Name)
	nameStyle := lipgloss.NewStyle().Foreground(p.Text).Bold(true)
	if state == "all" || state == "partial" {
		nameStyle = nameStyle.Foreground(p.Primary)
	}
	nameRendered := nameStyle.Render(name)
	nameCol := padName(nameRendered, name, 28)

	sel := m.bundleSelectedCount(bundle)
	total := len(bundle.Tools)
	statusText := fmt.Sprintf("%d tools · %d selected", total, sel)
	status := lipgloss.NewStyle().Foreground(p.Muted).Render(statusText)

	return fmt.Sprintf("%s %s %s %s", box, caretStyled, nameCol, status)
}

// renderToolRow: indented under parent. Checkbox sits flush with name —
// no connector glyph, no ID column. Indent = bundle-checkbox width so
// the tool checkbox visually nests one step in.
func (m treePickerModel) renderToolRow(row treeRow) string {
	p := m.palette
	tool := m.findTool(row.bundleID, row.toolName)
	selected := m.selected[row.bundleID+"/"+row.toolName]

	state := "none"
	if selected {
		state = "all"
	}
	box := checkbox(p, state)

	// 4-space indent: aligns the tool checkbox under the bundle's caret
	// position (after [✓] + space). Keeps state→name visually attached.
	indent := "    "

	nameStyle := lipgloss.NewStyle().Foreground(p.Text)
	if selected {
		nameStyle = nameStyle.Foreground(p.Primary).Bold(true)
	}
	nameRendered := nameStyle.Render(tool.Name)
	nameCol := padName(nameRendered, tool.Name, 22)

	var statePill string
	if tool.Installed {
		statePill = lipgloss.NewStyle().Foreground(p.Success).Render("v" + tool.Version)
	} else {
		statePill = lipgloss.NewStyle().Foreground(p.Muted).Italic(true).Render("missing")
	}
	stateCol := padPlain(statePill, 12)

	src := lipgloss.NewStyle().Foreground(p.Muted).Render(tool.Source)

	return fmt.Sprintf("%s%s %s  %s %s", indent, box, nameCol, stateCol, src)
}

// bundleID2Num returns the 1-based ordinal of a bundle within the slice.
func bundleID2Num(bundles []preset.Bundle, id string) int {
	for i, b := range bundles {
		if b.ID == id {
			return i + 1
		}
	}
	return 0
}

func padName(rendered, plain string, target int) string {
	w := lipgloss.Width(plain)
	if w >= target {
		return rendered
	}
	return rendered + strings.Repeat(" ", target-w)
}

func padPlain(rendered string, target int) string {
	w := lipgloss.Width(rendered)
	if w >= target {
		return rendered
	}
	return rendered + strings.Repeat(" ", target-w)
}
