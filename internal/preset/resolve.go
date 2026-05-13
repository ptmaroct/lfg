package preset

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ResolveVersion looks up the actual `latest` version a tool will land
// at install time by querying the right upstream registry. Returns the
// resolved version string (no leading "v"), or empty + error on
// failure. Best-effort — caller should keep showing the parsed
// PlannedVersion as fallback.
//
// Backends covered:
//   - npm-distributed packages (Source="npm" OR mise/curl tools that
//     happen to also be on npm: pnpm, bun, yarn, agent-browser)
//   - node-lts (special — nodejs.org/dist/index.json)
//   - brew formulas
//
// Skipped: curl bash-script installs (no canonical registry), python
// (mise resolves via its own version DB), opaque mise-only plugins.
//
// Caller passes a context with a tight (1-2s) timeout; we don't want
// the tree picker to stall on a slow registry.
func ResolveVersion(ctx context.Context, t Tool) (string, error) {
	switch t.Name {
	case "node-lts":
		return resolveNodeLTS(ctx)
	}

	// Try npm-registry-style lookup for anything that's likely on npm:
	// either Source="npm" explicitly, or a mise/curl tool whose name
	// matches a known npm package. We don't probe blindly — only the
	// allowlist below to avoid false hits on unrelated names.
	npmName, ok := npmPackageFor(t)
	if ok {
		return resolveNpm(ctx, npmName)
	}

	if t.Source == "brew" || t.Source == "cask" {
		return resolveBrew(ctx, t.Name)
	}

	return "", nil
}

// npmPackageFor returns the npm package id for tools we know to be
// distributed via npm. Some preset tools install via mise / curl but
// are also published on npm with the same name; we use npm as a
// reliable version oracle even when the chosen install path is
// different. Allow-list keeps the lookup honest.
func npmPackageFor(t Tool) (string, bool) {
	if t.Source == "npm" {
		// Pull the package off the install command (handles scoped
		// names like @openai/codex). Falls back to t.Name when the
		// command can't be parsed.
		cmd := t.InstallMac
		if cmd == "" {
			cmd = t.InstallLinux
		}
		if pkg := pkgFromNpmCmd(cmd); pkg != "" {
			return pkg, true
		}
		return t.Name, true
	}
	switch t.Name {
	case "pnpm", "bun", "yarn", "agent-browser":
		return t.Name, true
	}
	return "", false
}

func pkgFromNpmCmd(cmd string) string {
	fields := strings.Fields(cmd)
	for i, f := range fields {
		if f == "install" || f == "i" {
			if i+1 < len(fields) {
				next := fields[i+1]
				if next == "-g" || next == "--global" {
					if i+2 < len(fields) {
						return fields[i+2]
					}
				} else {
					return next
				}
			}
		}
	}
	return ""
}

func resolveNpm(ctx context.Context, pkg string) (string, error) {
	url := fmt.Sprintf("https://registry.npmjs.org/%s/latest", pkg)
	body, err := httpGet(ctx, url)
	if err != nil {
		return "", err
	}
	var out struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", fmt.Errorf("decode npm %s: %w", pkg, err)
	}
	return out.Version, nil
}

func resolveBrew(ctx context.Context, name string) (string, error) {
	url := fmt.Sprintf("https://formulae.brew.sh/api/formula/%s.json", name)
	body, err := httpGet(ctx, url)
	if err != nil {
		return "", err
	}
	var out struct {
		Versions struct {
			Stable string `json:"stable"`
		} `json:"versions"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", fmt.Errorf("decode brew %s: %w", name, err)
	}
	return out.Versions.Stable, nil
}

func resolveNodeLTS(ctx context.Context) (string, error) {
	body, err := httpGet(ctx, "https://nodejs.org/dist/index.json")
	if err != nil {
		return "", err
	}
	var releases []struct {
		Version string      `json:"version"` // "v20.11.0"
		LTS     interface{} `json:"lts"`     // false OR codename string
	}
	if err := json.Unmarshal(body, &releases); err != nil {
		return "", fmt.Errorf("decode nodejs index: %w", err)
	}
	for _, r := range releases {
		if s, ok := r.LTS.(string); ok && s != "" {
			return strings.TrimPrefix(r.Version, "v") + " (" + s + ")", nil
		}
	}
	return "", fmt.Errorf("no LTS release in nodejs index")
}

// ResolveLatestStable is the bumper-facing sibling of ResolveVersion.
// It returns the highest registry version whose publish timestamp is
// at least minAge old AND which is not a pre-release / deprecated.
// Empty string + nil error means "no eligible version" (e.g. brew
// formulas where we have no publish-date data — bumper just keeps the
// existing pin in that case).
func ResolveLatestStable(ctx context.Context, t Tool, minAge time.Duration) (string, error) {
	switch t.Name {
	case "node-lts":
		return resolveNodeLTSStable(ctx, minAge)
	}
	if npmName, ok := npmPackageFor(t); ok {
		return resolveNpmStable(ctx, npmName, minAge)
	}
	if t.Source == "brew" || t.Source == "cask" {
		// formulae.brew.sh doesn't expose release dates; we trust brew
		// for our quarantine purposes and return the current stable
		// version. Bumper still surfaces the diff for human review.
		return resolveBrew(ctx, t.Name)
	}
	return "", nil
}

// resolveNpmStable hits the full registry document (not /latest) so
// we can read the `time` map per version. Strategy:
//
//  1. Trust `dist-tags.latest` IF its publish ts is older than the
//     quarantine window. This respects maintainer intent (e.g. Yarn
//     keeps `latest` on the classic 1.x line; my-package-2.0 might
//     be tagged `next`).
//  2. Otherwise walk all non-deprecated, non-pre-release versions
//     descending and return the first one inside the quarantine.
func resolveNpmStable(ctx context.Context, pkg string, minAge time.Duration) (string, error) {
	body, err := httpGet(ctx, "https://registry.npmjs.org/"+pkg)
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
		return "", fmt.Errorf("decode npm %s: %w", pkg, err)
	}
	cutoff := time.Now().Add(-minAge)

	if latest, ok := doc.DistTags["latest"]; ok && latest != "" {
		if raw, ok2 := doc.Time[latest]; ok2 {
			if ts, err := time.Parse(time.RFC3339, raw); err == nil && !ts.After(cutoff) {
				meta := doc.Versions[latest]
				if meta.Deprecated == "" && !isPreRelease(latest) {
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
		return "", fmt.Errorf("no eligible version for %s within quarantine window", pkg)
	}
	sort.Slice(pool, func(i, j int) bool { return semverLess(pool[i].ver, pool[j].ver) })
	return pool[len(pool)-1].ver, nil
}

func resolveNodeLTSStable(ctx context.Context, minAge time.Duration) (string, error) {
	body, err := httpGet(ctx, "https://nodejs.org/dist/index.json")
	if err != nil {
		return "", err
	}
	var releases []struct {
		Version string      `json:"version"`
		Date    string      `json:"date"`
		LTS     interface{} `json:"lts"`
	}
	if err := json.Unmarshal(body, &releases); err != nil {
		return "", fmt.Errorf("decode nodejs index: %w", err)
	}
	cutoff := time.Now().Add(-minAge)
	for _, r := range releases {
		s, ok := r.LTS.(string)
		if !ok || s == "" {
			continue
		}
		ts, err := time.Parse("2006-01-02", r.Date)
		if err != nil {
			continue
		}
		if ts.After(cutoff) {
			continue
		}
		return strings.TrimPrefix(r.Version, "v"), nil
	}
	return "", fmt.Errorf("no LTS release older than quarantine window")
}

// isPreRelease flags semver strings with a hyphenated suffix (e.g.
// `1.2.3-alpha.0`, `0.5.0-rc.1`). Build metadata (`+`) is treated as
// stable.
func isPreRelease(v string) bool {
	v = strings.TrimPrefix(v, "v")
	return strings.Contains(v, "-")
}

// semverLess compares two MAJOR.MINOR.PATCH strings numerically.
// Non-numeric or short versions sort as zeros. Sufficient for our use
// (registry-published versions, no exotic forms).
func semverLess(a, b string) bool {
	pa, pb := parseSemver(a), parseSemver(b)
	for i := 0; i < 3; i++ {
		if pa[i] != pb[i] {
			return pa[i] < pb[i]
		}
	}
	return false
}

func parseSemver(s string) [3]int {
	s = strings.TrimPrefix(s, "v")
	if i := strings.IndexAny(s, "-+"); i >= 0 {
		s = s[:i]
	}
	var out [3]int
	parts := strings.SplitN(s, ".", 3)
	for i := 0; i < 3 && i < len(parts); i++ {
		out[i], _ = strconv.Atoi(parts[i])
	}
	return out
}

func httpGet(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "lfg/0.2 (+https://github.com/ptmaroct/lfg)")
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("%s: %s", url, resp.Status)
	}
	const maxRead = 16 << 20 // 16MB — full npm registry docs (e.g. bun) exceed 1MB
	limited := http.MaxBytesReader(nil, resp.Body, maxRead)
	buf := make([]byte, 0, 4096)
	chunk := make([]byte, 4096)
	for {
		n, err := limited.Read(chunk)
		if n > 0 {
			buf = append(buf, chunk[:n]...)
		}
		if err != nil {
			break
		}
	}
	return buf, nil
}
