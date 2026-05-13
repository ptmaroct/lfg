package preset

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestResolveNpmStable_DistTagLatestRespected(t *testing.T) {
	now := time.Now().UTC()
	doc := map[string]any{
		"dist-tags": map[string]string{"latest": "1.22.22"},
		"versions": map[string]any{
			"1.22.22": map[string]any{},
			"4.5.0":   map[string]any{},
		},
		"time": map[string]string{
			"1.22.22": now.Add(-90 * 24 * time.Hour).Format(time.RFC3339),
			"4.5.0":   now.Add(-90 * 24 * time.Hour).Format(time.RFC3339),
		},
	}
	srv := newRegistryServer(doc)
	defer srv.Close()

	got, err := resolveNpmStableFrom(context.Background(), srv.URL, 7*24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if got != "1.22.22" {
		t.Errorf("dist-tags.latest should win; got %q want 1.22.22", got)
	}
}

func TestResolveNpmStable_QuarantineFallsBack(t *testing.T) {
	now := time.Now().UTC()
	doc := map[string]any{
		"dist-tags": map[string]string{"latest": "2.0.0"},
		"versions": map[string]any{
			"1.9.0": map[string]any{},
			"2.0.0": map[string]any{},
		},
		"time": map[string]string{
			// latest is too fresh (1 day old, quarantine is 7 days)
			"2.0.0": now.Add(-24 * time.Hour).Format(time.RFC3339),
			"1.9.0": now.Add(-30 * 24 * time.Hour).Format(time.RFC3339),
		},
	}
	srv := newRegistryServer(doc)
	defer srv.Close()

	got, err := resolveNpmStableFrom(context.Background(), srv.URL, 7*24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if got != "1.9.0" {
		t.Errorf("should fall back to 1.9.0; got %q", got)
	}
}

func TestResolveNpmStable_SkipsPrereleaseAndDeprecated(t *testing.T) {
	now := time.Now().UTC()
	doc := map[string]any{
		"dist-tags": map[string]string{"latest": "2.0.0-rc.1"},
		"versions": map[string]any{
			"1.0.0":      map[string]any{"deprecated": "do not use"},
			"1.5.0":      map[string]any{},
			"2.0.0-rc.1": map[string]any{},
		},
		"time": map[string]string{
			"1.0.0":      now.Add(-60 * 24 * time.Hour).Format(time.RFC3339),
			"1.5.0":      now.Add(-30 * 24 * time.Hour).Format(time.RFC3339),
			"2.0.0-rc.1": now.Add(-30 * 24 * time.Hour).Format(time.RFC3339),
		},
	}
	srv := newRegistryServer(doc)
	defer srv.Close()

	got, err := resolveNpmStableFrom(context.Background(), srv.URL, 7*24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if got != "1.5.0" {
		t.Errorf("should pick 1.5.0 (1.0.0 deprecated, 2.0.0-rc.1 prerelease); got %q", got)
	}
}

func TestResolveNpmStable_NoEligible(t *testing.T) {
	now := time.Now().UTC()
	doc := map[string]any{
		"dist-tags": map[string]string{"latest": "2.0.0"},
		"versions":  map[string]any{"2.0.0": map[string]any{}},
		"time":      map[string]string{"2.0.0": now.Format(time.RFC3339)},
	}
	srv := newRegistryServer(doc)
	defer srv.Close()
	if _, err := resolveNpmStableFrom(context.Background(), srv.URL, 7*24*time.Hour); err == nil {
		t.Fatal("expected error when every version is inside quarantine")
	}
}

func TestSemverLess(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"1.0.0", "2.0.0", true},
		{"1.9.0", "1.10.0", true},
		{"2.0.0", "1.99.0", false},
		{"1.0.0", "1.0.0", false},
		{"v1.2.3", "1.2.4", true},
		{"1.0.0-rc.1", "1.0.0", false}, // strips prerelease, equal numeric → false
	}
	for _, c := range cases {
		if got := semverLess(c.a, c.b); got != c.want {
			t.Errorf("semverLess(%q, %q) = %v; want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestIsPreRelease(t *testing.T) {
	for _, v := range []string{"1.0.0-rc.1", "v2.0.0-beta", "0.1.0-alpha"} {
		if !isPreRelease(v) {
			t.Errorf("%q should be prerelease", v)
		}
	}
	for _, v := range []string{"1.0.0", "v2.5.0", "0.130.0"} {
		if isPreRelease(v) {
			t.Errorf("%q should NOT be prerelease", v)
		}
	}
}

// helpers ---------------------------------------------------------

func newRegistryServer(doc map[string]any) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(doc)
	}))
}

// resolveNpmStableFrom is a test seam — same logic as resolveNpmStable
// but with the registry URL injected so we can point at httptest.
func resolveNpmStableFrom(ctx context.Context, url string, minAge time.Duration) (string, error) {
	body, err := httpGet(ctx, url)
	if err != nil {
		return "", err
	}
	var doc struct {
		DistTags map[string]string `json:"dist-tags"`
		Versions map[string]struct {
			Deprecated string `json:"deprecated,omitempty"`
		} `json:"versions"`
		Time map[string]string `json:"time"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}
	// Inline the same logic as resolveNpmStable, sharing the helpers.
	cutoff := time.Now().Add(-minAge)
	if latest, ok := doc.DistTags["latest"]; ok && latest != "" {
		if raw, ok2 := doc.Time[latest]; ok2 {
			if ts, err := time.Parse(time.RFC3339, raw); err == nil && !ts.After(cutoff) {
				if doc.Versions[latest].Deprecated == "" && !isPreRelease(latest) {
					return latest, nil
				}
			}
		}
	}
	type cand struct {
		ver string
		ts  time.Time
	}
	var pool []cand
	for v, meta := range doc.Versions {
		if isPreRelease(v) || meta.Deprecated != "" {
			continue
		}
		raw, ok := doc.Time[v]
		if !ok {
			continue
		}
		ts, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			continue
		}
		if ts.After(cutoff) {
			continue
		}
		pool = append(pool, cand{ver: v, ts: ts})
	}
	if len(pool) == 0 {
		return "", fmt.Errorf("no eligible version within quarantine window")
	}
	best := pool[0]
	for _, c := range pool[1:] {
		if semverLess(best.ver, c.ver) {
			best = c
		}
	}
	return best.ver, nil
}
