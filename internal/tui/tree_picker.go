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
	kind     string // "bundle", "tool", or "subheader"
	bundleID string
	toolName string
	depth    int
	idx      int    // index within parent bundle (1-based for tools)
	label    string // for "subheader" rows: the label text
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
		// Skip already-installed tools — they shouldn't queue for re-install.
		bID, tName := splitKey(k)
		if t := findToolInBundles(bundles, bID, tName); t.Installed {
			continue
		}
		m.selected[k] = v
	}
	if len(initialTools) == 0 {
		for _, b := range bundles {
			if !initialBundleIDs[b.ID] {
				continue
			}
			for _, t := range b.Tools {
				if t.Installed {
					continue
				}
				m.selected[b.ID+"/"+t.Name] = true
			}
		}
	}
	// Mandatory tools that AREN'T installed yet stay always-on so they
	// can't be unchecked. Already-installed mandatory tools are simply
	// counted as installed and need no selection state.
	for _, b := range bundles {
		for _, t := range b.Tools {
			if t.Mandatory && !t.Installed {
				m.selected[b.ID+"/"+t.Name] = true
			}
		}
	}
	// All bundles start collapsed by default. Cleaner first impression.
	m.rebuildRows()
	return m
}

// splitKey decomposes "<bundleID>/<toolName>" → (bundleID, toolName).
func splitKey(k string) (string, string) {
	for i := 0; i < len(k); i++ {
		if k[i] == '/' {
			return k[:i], k[i+1:]
		}
	}
	return k, ""
}

// findToolInBundles is a free-function variant of findTool used during
// model construction (no receiver yet).
func findToolInBundles(bundles []preset.Bundle, bundleID, toolName string) preset.Tool {
	for _, b := range bundles {
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

func (m *treePickerModel) rebuildRows() {
	m.rows = m.rows[:0]
	nodeOK := m.nodeAvailable()
	for _, b := range m.bundles {
		m.rows = append(m.rows, treeRow{kind: "bundle", bundleID: b.ID})
		if !m.expanded[b.ID] {
			continue
		}
		// Skills section is gated on node — surface why before listing
		// tools so users understand the disabled state.
		if b.ID == "skills" && !nodeOK {
			m.rows = append(m.rows, treeRow{
				kind: "subheader", bundleID: b.ID,
				label: "REQUIRES node-lts — select it in barebones first",
			})
		}
		// Split tools into to-install and already-installed groups so
		// the user sees them as two distinct sections, not one mixed
		// list where they have to mentally filter "already done" rows.
		var toInstall, alreadyOK []preset.Tool
		for _, t := range b.Tools {
			if t.Installed {
				alreadyOK = append(alreadyOK, t)
			} else {
				toInstall = append(toInstall, t)
			}
		}
		for i, t := range toInstall {
			m.rows = append(m.rows, treeRow{
				kind: "tool", bundleID: b.ID, toolName: t.Name,
				depth: 1, idx: i + 1,
			})
		}
		if len(alreadyOK) > 0 {
			label := "ALREADY INSTALLED"
			if len(toInstall) == 0 {
				label = "ALREADY INSTALLED — nothing else to do here"
			}
			m.rows = append(m.rows, treeRow{
				kind: "subheader", bundleID: b.ID, label: label,
			})
			for i, t := range alreadyOK {
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

// bundleSelectionState only considers not-yet-installed tools.
// Already-installed tools are out of band — they don't participate in
// the selection state and shouldn't influence the bundle's checkbox.
func (m treePickerModel) bundleSelectionState(b preset.Bundle) string {
	on, off := 0, 0
	for _, t := range b.Tools {
		if t.Installed {
			continue
		}
		if m.selected[b.ID+"/"+t.Name] {
			on++
		} else {
			off++
		}
	}
	switch {
	case on+off == 0:
		// All tools in this bundle already installed.
		return "done"
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
		if t.Installed {
			continue
		}
		if m.selected[b.ID+"/"+t.Name] {
			n++
		}
	}
	return n
}

// bundleInstalledCount counts tools detect already found.
func (m treePickerModel) bundleInstalledCount(b preset.Bundle) int {
	n := 0
	for _, t := range b.Tools {
		if t.Installed {
			n++
		}
	}
	return n
}

// bundlePendingTotal — tools that aren't installed yet (the pool the
// checkbox + selection counter operates over).
func (m treePickerModel) bundlePendingTotal(b preset.Bundle) int {
	return len(b.Tools) - m.bundleInstalledCount(b)
}

func (m *treePickerModel) toggleAtCursor() {
	row := m.rows[m.cursor]
	if row.kind == "subheader" {
		return
	}
	// Skills bundle requires node-lts. Block toggling until node is
	// either already installed or queued in the install set; otherwise
	// `npx skills add` won't have a runtime to execute.
	if row.bundleID == "skills" && !m.nodeAvailable() {
		return
	}
	if row.kind == "tool" {
		tool := m.findTool(row.bundleID, row.toolName)
		if tool.Installed || tool.Mandatory {
			return // installed = nothing to do; mandatory = forced on
		}
		key := row.bundleID + "/" + row.toolName
		m.selected[key] = !m.selected[key]
		return
	}
	bundle := m.findBundle(row.bundleID)
	target := m.bundleSelectionState(bundle) != "all"
	for _, t := range bundle.Tools {
		if t.Installed {
			continue // installed tools never selected
		}
		if t.Mandatory && !target {
			continue // can't bulk-deselect mandatory rows
		}
		m.selected[bundle.ID+"/"+t.Name] = target
	}
}

// nodeAvailable reports whether node-lts is installed on the host or
// queued in the current selection set. Skills install via `npx skills
// add`, so node has to be in the user's path by the time the skills
// step runs.
func (m treePickerModel) nodeAvailable() bool {
	for _, b := range m.bundles {
		for _, t := range b.Tools {
			if t.Name != "node-lts" {
				continue
			}
			if t.Installed {
				return true
			}
			if m.selected[b.ID+"/"+t.Name] {
				return true
			}
		}
	}
	return false
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
	if row.kind != "bundle" {
		return
	}
	m.expanded[row.bundleID] = true
	m.rebuildRows()
	// Anchor the bundle to the top of the viewport so its children
	// scroll into view below it. Without this, expanding the last
	// bundle on screen leaves the cursor pinned at the bottom and the
	// user has to manually scroll to see anything they just revealed.
	m.offset = m.cursor
	// Move cursor to the first child so the next ↓ press steps through
	// the expanded contents naturally (mirrors how filesystem tree
	// pickers in most editors behave).
	for i := m.cursor + 1; i < len(m.rows); i++ {
		if m.rows[i].kind == "tool" && m.rows[i].bundleID == row.bundleID {
			m.cursor = i
			break
		}
	}
	if m.cursor >= m.offset+m.pageSize {
		m.offset = m.cursor - m.pageSize + 1
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
		case "i", "I":
			// Tool info overlay. Only meaningful on tool rows; bundle
			// rows just no-op so the key doesn't beep.
			row := m.rows[m.cursor]
			if row.kind == "tool" {
				tool := m.findTool(row.bundleID, row.toolName)
				return m, openInfoCmd(row.bundleID, tool)
			}
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
			// Enter always proceeds. Expand/collapse is reserved for
			// arrow keys (→ ←) so Enter doesn't change meaning based
			// on cursor row state — pressing it twice on a bundle
			// shouldn't first expand and then advance.
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
		case "esc", "backspace", "delete":
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
			KeyHint(p, "I", "info"),
			KeyHint(p, "A", "all"),
			KeyHint(p, "⏎", "next"),
			KeyHint(p, "⎋", "back"),
		),
		height < 22,
	)
}

// Strict grid widths shared by the header, bundle row, and tool row.
// Anything that prints inside one of these columns must use the same
// width so the columns visually line up regardless of row type.
const (
	gridGutterW = 2  // "▸ " cursor or "  " padding
	gridBoxW    = 3  // [✓] / [~] / [ ] / [●]
	gridSpace   = 1  // separator between box and caret
	gridCaretW  = 1  // ▶ ▼ (bundle) or "·" (tool)
	gridSpace2  = 1  // separator between caret and name
	gridNameW   = 22
	gridGap     = 2 // gap between name and INSTALLED column
	gridStatusW = 12
)

// gridLeading is the column offset before the NAME column starts.
// Used by the header to align with bundle/tool row content.
const gridLeading = gridGutterW + gridBoxW + gridSpace + gridCaretW + gridSpace2

func (m treePickerModel) renderColumnHeader(contentW int) string {
	p := m.palette
	style := lipgloss.NewStyle().Foreground(p.Muted)
	return strings.Repeat(" ", gridLeading) +
		style.Render(padRightPlain("BUNDLE / TOOL", gridNameW)) +
		strings.Repeat(" ", gridGap) +
		style.Render(padRightPlain("INSTALLED", gridStatusW)) +
		" " +
		style.Render("VIA")
}

// checkbox returns a clear, dumb-obvious selection glyph.
//
//	[ ] = none (Muted, not Subtle — Subtle blends into the background)
//	[~] = partial (Warn, bold)
//	[✓] = all / selected (Primary, bold)
//	[●] = mandatory (always-on, can't be toggled — high-contrast green)
//	[✓] = done — already installed (Muted, NOT bold; reads as background)
func checkbox(p Palette, state string) string {
	switch state {
	case "all":
		return lipgloss.NewStyle().Bold(true).Foreground(p.Primary).Render("[✓]")
	case "partial":
		return lipgloss.NewStyle().Bold(true).Foreground(p.Warn).Render("[~]")
	case "mandatory":
		return lipgloss.NewStyle().Bold(true).Foreground(p.Success).Render("[●]")
	case "done":
		// Subdued so installed-tool rows recede; the eye finds the
		// pickable rows first instead of the dense green grid.
		return lipgloss.NewStyle().Foreground(p.Muted).Render("[✓]")
	default:
		return lipgloss.NewStyle().Bold(true).Foreground(p.Muted).Render("[ ]")
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

	switch row.kind {
	case "bundle":
		return gutter + m.renderBundleRow(row)
	case "subheader":
		return gutter + m.renderSubheaderRow(row)
	default:
		return gutter + m.renderToolRow(row)
	}
}

// renderSubheaderRow — dim italic divider introducing the "ALREADY
// INSTALLED" group inside an expanded bundle. Aligned with the grid's
// NAME column so it sits cleanly above the rows it labels.
func (m treePickerModel) renderSubheaderRow(row treeRow) string {
	p := m.palette
	style := lipgloss.NewStyle().Foreground(p.Muted).Italic(true)
	// gridLeading - gridGutterW because gutter is added by renderRow.
	return strings.Repeat(" ", gridLeading-gridGutterW) + style.Render(row.label)
}

// renderBundleRow uses the strict shared grid:
//
//	[box] [caret] NAME(22)  STATUS(rest of line)
//
// Status fills the INSTALLED + VIA columns (it's a single string for
// bundle rows since they don't have per-tool version/source).
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
	switch state {
	case "all", "partial":
		nameStyle = nameStyle.Foreground(p.Primary)
	case "done":
		nameStyle = nameStyle.Foreground(p.Muted) // recede
	}
	nameRendered := nameStyle.Render(name)
	nameCol := padName(nameRendered, name, gridNameW)

	pending := m.bundlePendingTotal(bundle)
	installed := m.bundleInstalledCount(bundle)
	sel := m.bundleSelectedCount(bundle)
	total := len(bundle.Tools)

	var statusText string
	switch {
	case pending == 0:
		statusText = fmt.Sprintf("all %d done", total)
	case installed == 0:
		statusText = fmt.Sprintf("%d/%d picked", sel, pending)
	default:
		statusText = fmt.Sprintf("%d/%d picked  ·  %d done", sel, pending, installed)
	}
	status := lipgloss.NewStyle().Foreground(p.Muted).Render(statusText)

	// box + space + caret + space + name + gap + status (free-form, fills rest)
	return box + " " + caretStyled + " " + nameCol + strings.Repeat(" ", gridGap) + status
}

// renderToolRow uses the same strict grid as bundle rows so columns
// line up. Tools occupy the bundle's caret column with a thin "·"
// connector instead — visual nesting without a separate indent that
// would offset every column rightward.
//
// Columns: [box] · NAME(22)  CURRENT(12) VIA
//   - CURRENT = version installed now (— when missing)
//   - VIA     = installer backend (brew/mise/npm/curl/skills)
func (m treePickerModel) renderToolRow(row treeRow) string {
	p := m.palette
	tool := m.findTool(row.bundleID, row.toolName)
	selected := m.selected[row.bundleID+"/"+row.toolName]

	var box string
	switch {
	case tool.Installed:
		box = checkbox(p, "done")
	case tool.Mandatory:
		box = checkbox(p, "mandatory")
	case selected:
		box = checkbox(p, "all")
	default:
		box = checkbox(p, "none")
	}

	// "·" sits where bundle rows put the caret — keeps the column grid
	// stable and gives a subtle nesting cue without a 4-space indent.
	connector := lipgloss.NewStyle().Foreground(p.Subtle).Render("·")

	nameStyle := lipgloss.NewStyle().Foreground(p.Text)
	switch {
	case tool.Installed:
		nameStyle = nameStyle.Foreground(p.Muted)
	case selected:
		nameStyle = nameStyle.Foreground(p.Primary).Bold(true)
	}
	nameRendered := nameStyle.Render(tool.Name)
	nameCol := padName(nameRendered, tool.Name, gridNameW)

	var current string
	if tool.Installed {
		v := tool.Version
		if v == "" {
			v = "installed"
		} else {
			v = "v" + v
		}
		current = lipgloss.NewStyle().Foreground(p.Muted).Render(v)
	} else {
		current = lipgloss.NewStyle().Foreground(p.Subtle).Render("—")
	}
	currentCol := padPlain(current, gridStatusW)

	srcColor := p.Muted
	if tool.Installed {
		srcColor = p.Subtle
	}
	src := lipgloss.NewStyle().Foreground(srcColor).Render(tool.Source)

	// box + space + connector + space + name + gap + current + space + src
	return box + " " + connector + " " + nameCol +
		strings.Repeat(" ", gridGap) + currentCol + " " + src
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
