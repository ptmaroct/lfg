package detect

import "github.com/ptmaroct/lfg/internal/installer"

// ProbeAliases scans the user's shell rc files for pre-existing alias
// definitions outside the lfg-managed fenced block. Returned map is
// keyed by alias name with value `<rc-basename>:<lineno>` so the
// alias picker can render conflict warnings before the user selects
// an alias that would clash.
//
// Thin wrapper around installer.ExistingAliases — kept here so the
// probe-phase fan-out in tui/probe.go can import it alongside Probe /
// ProbeAll without dragging the installer package in directly.
func ProbeAliases() map[string]string {
	return installer.ExistingAliases()
}
