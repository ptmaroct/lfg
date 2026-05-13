package preset

import (
	"testing"
	"time"
)

func TestParsePins_RoundTrip(t *testing.T) {
	src := `bumped_at = "2026-05-13T12:00:00Z"

[pins."barebones/node-lts"]
version = "24.15.0"

[pins."barebones/brew"]
sha256 = "abc123"
`
	ps, err := ParsePins([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	want, _ := time.Parse(time.RFC3339, "2026-05-13T12:00:00Z")
	if !ps.BumpedAt.Equal(want) {
		t.Errorf("bumped_at: got %v want %v", ps.BumpedAt, want)
	}
	if got := ps.Pins["barebones/node-lts"].Version; got != "24.15.0" {
		t.Errorf("node-lts version: got %q", got)
	}
	if got := ps.Pins["barebones/brew"].SHA256; got != "abc123" {
		t.Errorf("brew sha: got %q", got)
	}
}

func TestApplyPins_MergesIntoBundles(t *testing.T) {
	bundles := []Bundle{
		{
			ID: "barebones",
			Tools: []Tool{
				{Name: "node-lts", Source: "mise", InstallMac: "mise use -g node@lts"},
				{Name: "yarn", Source: "npm", InstallMac: "npm install -g yarn"},
			},
		},
	}
	ps := PinSet{
		BumpedAt: time.Now(),
		Pins: map[string]PinEntry{
			"barebones/node-lts": {Version: "22.0.0"},
			"barebones/yarn":     {Version: "1.22.22"},
		},
	}
	out := applyPins(bundles, ps)
	if out[0].Tools[0].Pin != "22.0.0" {
		t.Errorf("node-lts pin not applied: %q", out[0].Tools[0].Pin)
	}
	if out[0].Tools[1].Pin != "1.22.22" {
		t.Errorf("yarn pin not applied: %q", out[0].Tools[1].Pin)
	}
	// Source untouched: bundles parameter still has empty Pin.
	if bundles[0].Tools[0].Pin != "" {
		t.Errorf("applyPins mutated input slice")
	}
}

func TestSetRemotePins_OnlyReplacesIfNewer(t *testing.T) {
	// reset for isolation
	pinsMu.Lock()
	pinsActive = PinSet{BumpedAt: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), Pins: map[string]PinEntry{"x": {Version: "1.0"}}}
	pinsLoaded = true
	pinsMu.Unlock()

	// Older — must be ignored.
	SetRemotePins(PinSet{
		BumpedAt: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		Pins:     map[string]PinEntry{"x": {Version: "0.9"}},
	})
	if got := CurrentPins().Pins["x"].Version; got != "1.0" {
		t.Errorf("older pins should not replace; got %q", got)
	}

	// Newer — accepted.
	SetRemotePins(PinSet{
		BumpedAt: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		Pins:     map[string]PinEntry{"x": {Version: "2.0"}},
	})
	if got := CurrentPins().Pins["x"].Version; got != "2.0" {
		t.Errorf("newer pins should replace; got %q", got)
	}

	// Empty — ignored.
	SetRemotePins(PinSet{BumpedAt: time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)})
	if got := CurrentPins().Pins["x"].Version; got != "2.0" {
		t.Errorf("empty pins should not replace; got %q", got)
	}
}

func TestPinFreshness_Buckets(t *testing.T) {
	cases := []struct {
		ageDays    int
		wantBucket string
	}{
		{5, "fresh"},
		{14, "fresh"},
		{20, "stale"},
		{30, "stale"},
		{45, "very-stale"},
	}
	for _, c := range cases {
		pinsMu.Lock()
		pinsActive = PinSet{BumpedAt: time.Now().Add(-time.Duration(c.ageDays) * 24 * time.Hour)}
		pinsLoaded = true
		pinsMu.Unlock()
		_, b := PinFreshness()
		if b != c.wantBucket {
			t.Errorf("age=%d: got %q want %q", c.ageDays, b, c.wantBucket)
		}
	}
}
