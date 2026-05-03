// Package tui wires the screen state machine for `lfg`.
package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ptmaroct/lfg/internal/preset"
)

// screen is the high-level TUI state enum.
type screen int

const (
	screenWelcome screen = iota
	screenTree    // unified bundle + tool picker (tree view)
	screenBundles
	screenTools
	screenConfirm
	screenCreds // MCP credentials wizard, between confirm and progress
	screenProgress
	screenAliases // optional shell aliases — between progress and done
	screenDone
	screenBackupPrompt
	screenBackupDone
	screenQuit
	screenQuitConfirm
	screenProbe       // first-paint detection screen; entered only via NewWithProbe
	screenConfigInput // welcome → "load config file" input dialog
	screenInfo        // tree picker → "i" key → tool metadata overlay
	screenExport      // welcome → "export this machine" save-to-TOML dialog
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
	creds        credsModel
	progress     progressModel
	aliases      aliasesModel
	done         doneModel
	backup       backupModel
	quitConfirm  quitConfirmModel
	probe        probeModel
	configInput  configInputModel
	info         infoModel
	infoPrev     screen // screen to return to after info dialog closes
	export       exportModel

	selectedBundleIDs map[string]bool
	selectedTools     map[string]bool // key = bundleID + "/" + toolName

	// Alias state. aliasGroups is the catalog (defaults or config-driven);
	// selectedAliases keys are <group>/<aliasName>; aliasConflicts maps
	// alias name → "<rc-basename>:<line>" for pre-existing definitions
	// outside the lfg-managed block (populated during the probe phase).
	aliasGroups     []preset.AliasGroup
	selectedAliases map[string]bool
	aliasConflicts  map[string]string

	// progressRunner is injected via WithProgressRunner; nil → mock.
	progressRunner ProgressRunner
}

// Option mutates a Model under construction. See WithProgressRunner.
type Option func(*Model)

// WithProgressRunner injects the function used to drive installs on the
// progress screen. Default is the mock runner (sleeps + canned output)
// so snapshot tests + the snap helper never touch the host system.
// CLI startup passes installer.Run for the real thing.
func WithProgressRunner(r ProgressRunner) Option {
	return func(m *Model) { m.progressRunner = r }
}

// New builds the root model with the given theme using the default
// (hardcoded) bundle data from the preset package. Snapshot tests rely
// on this constructor giving deterministic, machine-independent state.
//
// Real CLI startup uses NewWithBundles after running detect so the
// picker reflects what's actually installed.
func New(theme ThemeName) Model {
	return NewWithBundles(theme, preset.FilterForHost(preset.All()))
}

// NewWithBundles is the explicit constructor used after a detect pass.
// Pass the bundle slice with Installed/Version fields already overlaid
// (see internal/detect.Apply). Defaults aliases to the built-in
// catalog with no conflict map; callers wanting config-driven aliases
// or pre-probed conflicts should use NewWithBundlesAndAliases.
func NewWithBundles(theme ThemeName, bundles []preset.Bundle, opts ...Option) Model {
	return NewWithBundlesAndAliases(theme, bundles, preset.DefaultAliases(), nil, opts...)
}

// NewWithBundlesAndAliases is the full constructor — bundles + alias
// catalog + alias conflict map (probed from rc files). Callers driven
// by --config thread their parsed catalog through here so the alias
// picker honors the config as source of truth (REPLACE semantics, same
// as bundle precedence).
func NewWithBundlesAndAliases(theme ThemeName, bundles []preset.Bundle, aliasGroups []preset.AliasGroup, conflicts map[string]string, opts ...Option) Model {
	p := PaletteFor(theme)
	pre := map[string]bool{}
	for _, bundle := range bundles {
		if bundle.Default {
			pre[bundle.ID] = true
		}
	}
	m := Model{
		screen:            screenWelcome,
		palette:           p,
		theme:             theme,
		bundles:           bundles,
		selectedBundleIDs: pre,
		selectedTools:     map[string]bool{},
		aliasGroups:       aliasGroups,
		aliasConflicts:    conflicts,
	}
	m.welcome = newWelcome(p)
	for _, o := range opts {
		o(&m)
	}
	return m
}

// NewWithProbe is the CLI-startup constructor. It takes raw (un-probed)
// bundles and starts on screenProbe so the user sees detection progress
// instead of a frozen terminal. The probe screen runs detect.ProbeAll
// in the background, then transitions to welcome with the result-applied
// bundles. Snapshot tests intentionally use New / NewWithBundles so they
// skip the probe step (deterministic state). Alias catalog defaults to
// the built-in set; use NewWithProbeAndAliases for config-driven catalog.
func NewWithProbe(theme ThemeName, raw []preset.Bundle, opts ...Option) Model {
	return NewWithProbeAndAliases(theme, raw, preset.DefaultAliases(), opts...)
}

// NewWithProbeAndAliases lets callers thread a custom alias catalog
// through the probe-first startup path. Alias conflicts get probed
// alongside detect and arrive via probeAliasesMsg.
func NewWithProbeAndAliases(theme ThemeName, raw []preset.Bundle, aliasGroups []preset.AliasGroup, opts ...Option) Model {
	p := PaletteFor(theme)
	pre := map[string]bool{}
	for _, bundle := range raw {
		if bundle.Default {
			pre[bundle.ID] = true
		}
	}
	m := Model{
		screen:            screenProbe,
		palette:           p,
		theme:             theme,
		bundles:           raw,
		selectedBundleIDs: pre,
		selectedTools:     map[string]bool{},
		aliasGroups:       aliasGroups,
	}
	m.probe = newProbe(p, raw)
	for _, o := range opts {
		o(&m)
	}
	return m
}

func (m Model) Init() tea.Cmd {
	if m.screen == screenProbe {
		return m.probe.Init()
	}
	return m.welcome.Init()
}

// Theme returns the currently active theme name. Used by the CLI layer
// to persist theme changes (Ctrl+T cycling) on clean exit.
func (m Model) Theme() ThemeName { return m.theme }

// SelectedAliases returns the resolved alias slice the user picked,
// in catalog order. Used by the CLI layer to write the lfg-managed
// alias block to shell rc files after the TUI exits.
func (m Model) SelectedAliases() []preset.Alias {
	return resolveSelectedAliases(m.aliasGroups, m.selectedAliases)
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
		// Theme cycle. Bound to multiple keys because terminals
		// vary on whether ctrl+t reaches the app (Ghostty + tmux
		// sometimes intercept it). Capital T is a reliable fallback
		// that won't collide with picker filter input (which uses
		// lowercase t).
		if msg.String() == "ctrl+t" || msg.String() == "T" {
			m.theme = nextTheme(m.theme)
			m.palette = PaletteFor(m.theme)
			return m.rehydrate()
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
	case openInfoMsg:
		// Stash whatever screen we're on so Esc/Enter inside info can return.
		m.infoPrev = m.screen
		m.info = newInfo(m.palette, msg.bundleID, msg.tool, m.screen)
		m.screen = screenInfo
		return m, m.info.Init()
	case closeInfoMsg:
		// Return to the screen we came from without rebuilding it —
		// preserves tree-picker cursor + per-screen state.
		m.screen = m.infoPrev
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
	case screenCreds:
		m.creds, cmd = m.creds.Update(msg)
	case screenProgress:
		m.progress, cmd = m.progress.Update(msg)
	case screenAliases:
		m.aliases, cmd = m.aliases.Update(msg)
	case screenDone:
		m.done, cmd = m.done.Update(msg)
	case screenBackupPrompt, screenBackupDone:
		m.backup, cmd = m.backup.Update(msg)
	case screenQuitConfirm:
		m.quitConfirm, cmd = m.quitConfirm.Update(msg)
	case screenProbe:
		m.probe, cmd = m.probe.Update(msg)
	case screenConfigInput:
		m.configInput, cmd = m.configInput.Update(msg)
	case screenInfo:
		m.info, cmd = m.info.Update(msg)
	case screenExport:
		m.export, cmd = m.export.Update(msg)
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
	case screenCreds:
		m.creds, cmd = m.creds.Update(msg)
	case screenProgress:
		m.progress, cmd = m.progress.Update(msg)
	case screenAliases:
		m.aliases, cmd = m.aliases.Update(msg)
	case screenDone:
		m.done, cmd = m.done.Update(msg)
	case screenBackupPrompt, screenBackupDone:
		m.backup, cmd = m.backup.Update(msg)
	case screenQuitConfirm:
		m.quitConfirm, cmd = m.quitConfirm.Update(msg)
	case screenProbe:
		m.probe, cmd = m.probe.Update(msg)
	case screenConfigInput:
		m.configInput, cmd = m.configInput.Update(msg)
	case screenInfo:
		m.info, cmd = m.info.Update(msg)
	case screenExport:
		m.export, cmd = m.export.Update(msg)
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
	case screenCreds:
		return m.creds.View(m.width, m.height)
	case screenProgress:
		return m.progress.View(m.width, m.height)
	case screenAliases:
		return m.aliases.View(m.width, m.height)
	case screenDone:
		return m.done.View(m.width, m.height)
	case screenBackupPrompt, screenBackupDone:
		return m.backup.View(m.width, m.height)
	case screenQuitConfirm:
		return m.quitConfirm.View(m.width, m.height)
	case screenProbe:
		return m.probe.View(m.width, m.height)
	case screenConfigInput:
		return m.configInput.View(m.width, m.height)
	case screenInfo:
		return m.info.View(m.width, m.height)
	case screenExport:
		return m.export.View(m.width, m.height)
	case screenQuit:
		return ""
	}
	return ""
}

// openInfoMsg pops the info overlay for a tool, remembering the screen
// to return to. Carried by openInfoCmd from the tree picker.
type openInfoMsg struct {
	bundleID string
	tool     preset.Tool
}

func openInfoCmd(bundleID string, t preset.Tool) tea.Cmd {
	return func() tea.Msg { return openInfoMsg{bundleID: bundleID, tool: t} }
}

// transitionMsg requests a screen change, optionally carrying state.
// `bundles` lets the probe screen replace the root bundle slice with
// detect-applied data when handing off to welcome. `selectedAliases`
// flows out of the alias picker; `aliasConflicts` flows in from the
// probe phase. `aliasGroups` lets a runtime --config load (from the
// welcome screen) swap the alias catalog mid-session.
type transitionMsg struct {
	target            screen
	selectedBundleIDs map[string]bool
	selectedTools     map[string]bool
	selectedAliases   map[string]bool
	bundles           []preset.Bundle
	aliasGroups       []preset.AliasGroup
	aliasConflicts    map[string]string
	replaceAliases    bool // when true, aliasGroups replaces (even if empty)
}

func (m Model) transition(msg transitionMsg) (tea.Model, tea.Cmd) {
	if msg.selectedBundleIDs != nil {
		m.selectedBundleIDs = msg.selectedBundleIDs
	}
	if msg.selectedTools != nil {
		m.selectedTools = msg.selectedTools
	}
	if msg.selectedAliases != nil {
		m.selectedAliases = msg.selectedAliases
	}
	if msg.aliasConflicts != nil {
		m.aliasConflicts = msg.aliasConflicts
	}
	if msg.replaceAliases {
		m.aliasGroups = msg.aliasGroups
		// Reset prior selection when the catalog changes — the keys
		// (group/name) likely don't carry over to the new catalog.
		m.selectedAliases = nil
	}
	if msg.bundles != nil {
		m.bundles = msg.bundles
		// Re-derive default bundle pre-selection from the freshly probed
		// data so the tree picker reflects current state.
		if len(m.selectedBundleIDs) == 0 {
			pre := map[string]bool{}
			for _, b := range m.bundles {
				if b.Default {
					pre[b.ID] = true
				}
			}
			m.selectedBundleIDs = pre
		}
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
	case screenCreds:
		m.creds = newCreds(m.palette, m.bundles, m.selectedTools)
		return m, m.creds.Init()
	case screenProgress:
		runner := m.progressRunner
		if runner == nil {
			runner = mockProgressRunner
		}
		m.progress = newProgressWithRunner(m.palette, m.bundles, m.selectedTools, runner)
		return m, m.progress.Init()
	case screenAliases:
		m.aliases = newAliases(m.palette, m.aliasGroups, m.selectedAliases, m.aliasConflicts)
		return m, m.aliases.Init()
	case screenDone:
		m.done = newDone(m.palette, m.bundles, m.selectedTools, resolveSelectedAliases(m.aliasGroups, m.selectedAliases))
		return m, m.done.Init()
	case screenBackupPrompt:
		m.backup = newBackup(m.palette)
		return m, m.backup.Init()
	case screenQuitConfirm:
		m.quitConfirm = newQuitConfirm(m.palette)
		return m, m.quitConfirm.Init()
	case screenProbe:
		m.probe = newProbe(m.palette, m.bundles)
		return m, m.probe.Init()
	case screenConfigInput:
		m.configInput = newConfigInput(m.palette)
		return m, m.configInput.Init()
	case screenInfo:
		// Restored after info dialog closes; the model already exists.
		return m, nil
	case screenExport:
		m.export = newExport(m.palette, m.bundles, flattenGroups(m.aliasGroups))
		return m, m.export.Init()
	case screenQuit:
		return m, tea.Quit
	}
	return m, nil
}

// nextTheme cycles through the built-in themes. Colorblind sits first
// in the cycle (after the default LFG theme) so accessibility users hit
// it on the very first Ctrl+T press without hunting through the menu.
//
// Order: lfg → colorblind → dracula → catppuccin → lfg
func nextTheme(t ThemeName) ThemeName {
	switch t {
	case ThemeLFG:
		return ThemeColorblind
	case ThemeColorblind:
		return ThemeDracula
	case ThemeDracula:
		return ThemeCatppuccin
	default:
		return ThemeLFG
	}
}

// rehydrate rebuilds the active child model with the current palette so
// a live theme swap (`t` key) takes effect immediately. Selection state
// flows through the root maps, so no data loss — only screen-local
// cursor / animation phase resets.
func (m Model) rehydrate() (tea.Model, tea.Cmd) {
	switch m.screen {
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
	case screenCreds:
		m.creds = newCreds(m.palette, m.bundles, m.selectedTools)
		return m, m.creds.Init()
	case screenProgress:
		runner := m.progressRunner
		if runner == nil {
			runner = mockProgressRunner
		}
		m.progress = newProgressWithRunner(m.palette, m.bundles, m.selectedTools, runner)
		return m, m.progress.Init()
	case screenAliases:
		m.aliases = newAliases(m.palette, m.aliasGroups, m.selectedAliases, m.aliasConflicts)
		return m, m.aliases.Init()
	case screenDone:
		m.done = newDone(m.palette, m.bundles, m.selectedTools, resolveSelectedAliases(m.aliasGroups, m.selectedAliases))
		return m, m.done.Init()
	case screenBackupPrompt, screenBackupDone:
		m.backup = newBackup(m.palette)
		return m, m.backup.Init()
	case screenQuitConfirm:
		m.quitConfirm = newQuitConfirm(m.palette)
		return m, m.quitConfirm.Init()
	case screenProbe:
		// Re-init would restart the probe goroutine and double-fire
		// detect work. The current probe instance is fine; only the
		// palette references are stale, and the next paint picks them
		// up on its own. So just keep the existing model.
		return m, nil
	case screenConfigInput:
		m.configInput = newConfigInput(m.palette)
		return m, m.configInput.Init()
	case screenInfo:
		m.info = newInfo(m.palette, m.info.bundleID, m.info.tool, m.infoPrev)
		return m, m.info.Init()
	case screenExport:
		m.export = newExport(m.palette, m.bundles, flattenGroups(m.aliasGroups))
		return m, m.export.Init()
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
