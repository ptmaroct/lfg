// Package tui wires the screen state machine for `lfg`.
package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/anuj/lfg/internal/preset"
)

// screen is the high-level TUI state enum.
type screen int

const (
	screenWelcome screen = iota
	screenTree    // unified bundle + tool picker (tree view)
	screenBundles
	screenTools
	screenConfirm
	screenProgress
	screenDone
	screenBackupPrompt
	screenBackupDone
	screenQuit
	screenQuitConfirm
)

// Model is the root bubbletea model. Each screen is a child model;
// Update dispatches based on current `screen`.
type Model struct {
	screen     screen
	prevScreen screen // saved before transitioning to screenQuitConfirm
	width      int
	height     int
	palette    Palette
	theme      ThemeName
	bundles    []preset.Bundle

	welcome      welcomeModel
	tree         treePickerModel
	bundlePicker bundlePickerModel
	toolPicker   toolPickerModel
	confirm      confirmModel
	progress     progressModel
	done         doneModel
	backup       backupModel
	quitConfirm  quitConfirmModel

	selectedBundleIDs map[string]bool
	selectedTools     map[string]bool // key = bundleID + "/" + toolName
}

// New builds the root model with the given theme.
func New(theme ThemeName) Model {
	p := PaletteFor(theme)
	bs := preset.All()
	pre := map[string]bool{}
	for _, bundle := range bs {
		if bundle.Default {
			pre[bundle.ID] = true
		}
	}
	m := Model{
		screen:            screenWelcome,
		palette:           p,
		theme:             theme,
		bundles:           bs,
		selectedBundleIDs: pre,
		selectedTools:     map[string]bool{},
	}
	m.welcome = newWelcome(p)
	return m
}

func (m Model) Init() tea.Cmd {
	return m.welcome.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m.forwardSize(msg)
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.screen = screenQuit
			return m, tea.Quit
		}
		// Global `q` → quit confirm dialog. Don't intercept if already in
		// the confirm screen (so Y/N letter keys reach the dialog).
		if msg.String() == "q" && m.screen != screenQuitConfirm {
			m.prevScreen = m.screen
			m.screen = screenQuitConfirm
			m.quitConfirm = newQuitConfirm(m.palette)
			return m, m.quitConfirm.Init()
		}
	case transitionMsg:
		return m.transition(msg)
	case quitCancelMsg:
		m.screen = m.prevScreen
		// Re-init the previous screen if it was welcome (so animation tick resumes).
		if m.screen == screenWelcome {
			return m, m.welcome.Init()
		}
		return m, nil
	}

	var cmd tea.Cmd
	switch m.screen {
	case screenWelcome:
		m.welcome, cmd = m.welcome.Update(msg)
	case screenTree:
		m.tree, cmd = m.tree.Update(msg)
	case screenBundles:
		m.bundlePicker, cmd = m.bundlePicker.Update(msg)
	case screenTools:
		m.toolPicker, cmd = m.toolPicker.Update(msg)
	case screenConfirm:
		m.confirm, cmd = m.confirm.Update(msg)
	case screenProgress:
		m.progress, cmd = m.progress.Update(msg)
	case screenDone:
		m.done, cmd = m.done.Update(msg)
	case screenBackupPrompt, screenBackupDone:
		m.backup, cmd = m.backup.Update(msg)
	case screenQuitConfirm:
		m.quitConfirm, cmd = m.quitConfirm.Update(msg)
	}
	return m, cmd
}

func (m Model) forwardSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.screen {
	case screenWelcome:
		m.welcome, cmd = m.welcome.Update(msg)
	case screenTree:
		m.tree, cmd = m.tree.Update(msg)
	case screenBundles:
		m.bundlePicker, cmd = m.bundlePicker.Update(msg)
	case screenTools:
		m.toolPicker, cmd = m.toolPicker.Update(msg)
	case screenConfirm:
		m.confirm, cmd = m.confirm.Update(msg)
	case screenProgress:
		m.progress, cmd = m.progress.Update(msg)
	case screenDone:
		m.done, cmd = m.done.Update(msg)
	case screenBackupPrompt, screenBackupDone:
		m.backup, cmd = m.backup.Update(msg)
	case screenQuitConfirm:
		m.quitConfirm, cmd = m.quitConfirm.Update(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	switch m.screen {
	case screenWelcome:
		return m.welcome.View(m.width, m.height)
	case screenTree:
		return m.tree.View(m.width, m.height)
	case screenBundles:
		return m.bundlePicker.View(m.width, m.height)
	case screenTools:
		return m.toolPicker.View(m.width, m.height)
	case screenConfirm:
		return m.confirm.View(m.width, m.height)
	case screenProgress:
		return m.progress.View(m.width, m.height)
	case screenDone:
		return m.done.View(m.width, m.height)
	case screenBackupPrompt, screenBackupDone:
		return m.backup.View(m.width, m.height)
	case screenQuitConfirm:
		return m.quitConfirm.View(m.width, m.height)
	case screenQuit:
		return ""
	}
	return ""
}

// transitionMsg requests a screen change, optionally carrying state.
type transitionMsg struct {
	target            screen
	selectedBundleIDs map[string]bool
	selectedTools     map[string]bool
}

func (m Model) transition(msg transitionMsg) (tea.Model, tea.Cmd) {
	if msg.selectedBundleIDs != nil {
		m.selectedBundleIDs = msg.selectedBundleIDs
	}
	if msg.selectedTools != nil {
		m.selectedTools = msg.selectedTools
	}
	m.screen = msg.target

	switch msg.target {
	case screenWelcome:
		m.welcome = newWelcome(m.palette)
		return m, m.welcome.Init()
	case screenTree:
		m.tree = newTreePicker(m.palette, m.bundles, m.selectedBundleIDs, m.selectedTools)
		return m, m.tree.Init()
	case screenBundles:
		m.bundlePicker = newBundlePicker(m.palette, m.bundles, m.selectedBundleIDs)
		return m, m.bundlePicker.Init()
	case screenTools:
		m.toolPicker = newToolPicker(m.palette, m.bundles, m.selectedBundleIDs, m.selectedTools)
		return m, m.toolPicker.Init()
	case screenConfirm:
		m.confirm = newConfirm(m.palette, m.bundles, m.selectedTools)
		return m, m.confirm.Init()
	case screenProgress:
		m.progress = newProgress(m.palette, m.bundles, m.selectedTools)
		return m, m.progress.Init()
	case screenDone:
		m.done = newDone(m.palette)
		return m, m.done.Init()
	case screenBackupPrompt:
		m.backup = newBackup(m.palette)
		return m, m.backup.Init()
	case screenQuitConfirm:
		m.quitConfirm = newQuitConfirm(m.palette)
		return m, m.quitConfirm.Init()
	case screenQuit:
		return m, tea.Quit
	}
	return m, nil
}

func goTo(target screen) tea.Cmd {
	return func() tea.Msg { return transitionMsg{target: target} }
}

func goToWithBundles(target screen, bundleIDs map[string]bool) tea.Cmd {
	return func() tea.Msg {
		return transitionMsg{target: target, selectedBundleIDs: bundleIDs}
	}
}

func goToWithTools(target screen, tools map[string]bool) tea.Cmd {
	return func() tea.Msg {
		return transitionMsg{target: target, selectedTools: tools}
	}
}
