package preset

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
)

// PinEntry is a single tool's pinned version + optional SHA256 (used
// when the install path pipes a remote script into a shell). Empty
// fields are permitted: a tool may have a Version pin without a SHA
// (npm/mise) or a SHA without a Version (pure curl installer where
// the script self-describes the version it'll install).
type PinEntry struct {
	Version string `toml:"version,omitempty"`
	SHA256  string `toml:"sha256,omitempty"`
}

// PinSet is the parsed view of pins.toml — the BumpedAt timestamp plus
// a map keyed by "<bundleID>/<toolName>".
type PinSet struct {
	BumpedAt time.Time           `toml:"bumped_at"`
	Pins     map[string]PinEntry `toml:"pins"`
}

//go:embed pins.toml
var embeddedPinsTOML []byte

var (
	pinsMu     sync.RWMutex
	pinsActive PinSet
	pinsLoaded bool
	// nowFunc is the clock used by PinFreshness. Tests override it via
	// SetNowForTest so snapshot output stays deterministic regardless
	// of when the suite runs.
	nowFunc = time.Now
)

// SetNowForTest pins the clock used by PinFreshness. Returns a func
// the caller defers to restore the real clock. Test-only — production
// code never calls this.
func SetNowForTest(t time.Time) func() {
	pinsMu.Lock()
	defer pinsMu.Unlock()
	prev := nowFunc
	nowFunc = func() time.Time { return t }
	return func() {
		pinsMu.Lock()
		defer pinsMu.Unlock()
		nowFunc = prev
	}
}

// SetPinsForTest replaces the active pin set unconditionally. Unlike
// SetRemotePins which only accepts newer sets, this is for snapshot
// tests that need a stable BumpedAt regardless of what's embedded.
// Returns a restore func.
func SetPinsForTest(ps PinSet) func() {
	pinsMu.Lock()
	defer pinsMu.Unlock()
	prev := pinsActive
	prevLoaded := pinsLoaded
	pinsActive = ps
	pinsLoaded = true
	return func() {
		pinsMu.Lock()
		defer pinsMu.Unlock()
		pinsActive = prev
		pinsLoaded = prevLoaded
	}
}

// CurrentPins returns the active pin set. On first call it parses the
// embedded copy; later calls return whatever was last set via
// SetRemotePins (when the TUI's async remote fetcher landed a fresher
// signed copy from GitHub raw). Falls back to embedded on any parse
// failure so the binary stays usable.
func CurrentPins() PinSet {
	pinsMu.RLock()
	if pinsLoaded {
		ps := pinsActive
		pinsMu.RUnlock()
		return ps
	}
	pinsMu.RUnlock()

	pinsMu.Lock()
	defer pinsMu.Unlock()
	if pinsLoaded {
		return pinsActive
	}
	ps, err := ParsePins(embeddedPinsTOML)
	if err != nil {
		// embedded file is checked-in by the repo — if it fails to
		// parse we just go with an empty set so the rest of the app
		// keeps working with un-pinned commands.
		ps = PinSet{}
	}
	pinsActive = ps
	pinsLoaded = true
	return pinsActive
}

// SetRemotePins installs a freshly-fetched pin set (e.g. from
// raw.githubusercontent.com/<repo>/main/internal/preset/pins.toml).
// Caller is responsible for any signature / freshness validation
// before calling — we only check that the new set is non-empty and
// newer than what we currently have.
func SetRemotePins(ps PinSet) {
	pinsMu.Lock()
	defer pinsMu.Unlock()
	if len(ps.Pins) == 0 {
		return
	}
	if pinsLoaded && !ps.BumpedAt.After(pinsActive.BumpedAt) {
		return
	}
	pinsActive = ps
	pinsLoaded = true
}

// ParsePins decodes a pins.toml byte buffer.
func ParsePins(data []byte) (PinSet, error) {
	var ps PinSet
	if err := toml.Unmarshal(data, &ps); err != nil {
		return PinSet{}, fmt.Errorf("parse pins.toml: %w", err)
	}
	if ps.Pins == nil {
		ps.Pins = map[string]PinEntry{}
	}
	return ps, nil
}

// EncodePins renders a PinSet back to TOML. Used by the preset-bump
// bot to rewrite the file on disk.
func EncodePins(ps PinSet) ([]byte, error) {
	return toml.Marshal(ps)
}

// applyPins walks raw bundles + copies Version / SHA256 from the pin
// set onto each Tool. Returns a new slice (caller-owned copy) so the
// rawBundles literal stays untouched between calls.
func applyPins(bundles []Bundle, ps PinSet) []Bundle {
	out := make([]Bundle, len(bundles))
	for i, b := range bundles {
		tools := make([]Tool, len(b.Tools))
		for j, t := range b.Tools {
			key := b.ID + "/" + t.Name
			if p, ok := ps.Pins[key]; ok {
				t.Pin = p.Version
				t.PinSHA256 = p.SHA256
			}
			tools[j] = t
		}
		b.Tools = tools
		out[i] = b
	}
	return out
}

// DefaultRemotePinsURL points at the main-branch pins.toml in the
// canonical repo. The TUI's async fetcher hits this URL on launch
// (2s timeout, fail-closed) so a binary released a month ago can
// still install today's recommended versions without `lfg update`.
const DefaultRemotePinsURL = "https://raw.githubusercontent.com/ptmaroct/lfg/main/internal/preset/pins.toml"

// FetchRemotePins downloads pins.toml from a URL and returns the
// parsed PinSet. HTTPS-only, 2s timeout, body capped at 256 KiB
// (current pins.toml is ~2 KiB). Caller decides what to do with the
// result; SetRemotePins won't accept it if it's older than what's
// already active.
//
// Note: signature verification (cosign) is intentionally deferred —
// this fetch is read-only, the result only affects in-process pin
// data, and a tampered file would surface immediately because the
// install step's SHA256 check (see runVerifiedCurl) would still fail.
// Cosign + .sig handling will land in a follow-up.
func FetchRemotePins(ctx context.Context, url string) (PinSet, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return PinSet{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "lfg-pins-fetcher (+https://github.com/ptmaroct/lfg)")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return PinSet{}, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return PinSet{}, fmt.Errorf("fetch %s: %s", url, resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 256<<10))
	if err != nil {
		return PinSet{}, fmt.Errorf("read body: %w", err)
	}
	return ParsePins(body)
}

// PinFreshness reports the age of the active pin set in whole days
// and a human-readable colour bucket. Used by the TUI welcome chrome
// to surface staleness — see docs/versioning.md.
func PinFreshness() (ageDays int, bucket string) {
	ps := CurrentPins()
	if ps.BumpedAt.IsZero() {
		return -1, "unknown"
	}
	age := nowFunc().Sub(ps.BumpedAt)
	d := int(age.Hours() / 24)
	switch {
	case d <= 14:
		return d, "fresh"
	case d <= 30:
		return d, "stale"
	default:
		return d, "very-stale"
	}
}
