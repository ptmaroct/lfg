// Command preset-bump regenerates internal/preset/pins.toml from
// upstream registries, applying the configured quarantine window so
// freshly-published versions can't slip into the lfg default preset
// before they've had time to be flagged. Run weekly from the
// preset-bump GitHub Action, but also usable locally:
//
//	go run ./cmd/preset-bump --min-age-days=7 --dry-run
//
// Exit codes:
//
//	0 — pins.toml written (or already up to date in --dry-run)
//	1 — fatal error
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/ptmaroct/lfg/internal/preset"
)

var urlRE = regexp.MustCompile(`https?://[^\s"'|)>]+`)

func main() {
	minAgeDays := flag.Int("min-age-days", 7, "minimum age in days before a new version is eligible")
	dryRun := flag.Bool("dry-run", false, "print diff but don't write pins.toml")
	output := flag.String("output", "internal/preset/pins.toml", "path to pins.toml")
	flag.Parse()

	ctx := context.Background()
	minAge := time.Duration(*minAgeDays) * 24 * time.Hour

	current := preset.CurrentPins()
	next := preset.PinSet{
		BumpedAt: time.Now().UTC(),
		Pins:     map[string]preset.PinEntry{},
	}

	type diffRow struct {
		key, oldVer, newVer, oldSHA, newSHA, note string
	}
	var rows []diffRow

	for _, b := range rawBundlesForBump() {
		for _, t := range b.Tools {
			// Skills install via `npx skills add` — there is no
			// per-tool version pin to track. Skip the bundle entirely.
			if t.Source == "skills" {
				continue
			}
			key := b.ID + "/" + t.Name
			cur := current.Pins[key]
			entry := preset.PinEntry{Version: cur.Version, SHA256: cur.SHA256}

			// Version resolution.
			if v, err := preset.ResolveLatestStable(ctx, t, minAge); err == nil && v != "" {
				entry.Version = v
			} else if err != nil {
				fmt.Fprintf(os.Stderr, "warn: resolve %s: %v\n", key, err)
			}

			// SHA256 drift for curl-piped install scripts. Check both
			// Mac and Linux commands; a curl-bash one-liner in either
			// counts. Heuristic: command must mention `curl` AND have
			// a pipe / command-substitution so we only hash actual
			// install scripts.
			note := ""
			for _, cmd := range []string{t.InstallMac, t.InstallLinux} {
				if !strings.Contains(cmd, "curl") {
					continue
				}
				if !strings.Contains(cmd, "|") && !strings.Contains(cmd, "$(") {
					continue
				}
				url := urlRE.FindString(cmd)
				if url == "" {
					continue
				}
				sha, err := fetchSHA(ctx, url)
				if err != nil {
					fmt.Fprintf(os.Stderr, "warn: sha %s: %v\n", url, err)
					break
				}
				if cur.SHA256 != "" && cur.SHA256 != sha {
					note = "⚠ install script CONTENT CHANGED — review before merge"
				}
				entry.SHA256 = sha
				break
			}

			if entry.Version == "" && entry.SHA256 == "" {
				continue
			}
			next.Pins[key] = entry
			if cur.Version != entry.Version || cur.SHA256 != entry.SHA256 || note != "" {
				rows = append(rows, diffRow{
					key: key, oldVer: cur.Version, newVer: entry.Version,
					oldSHA: cur.SHA256, newSHA: entry.SHA256, note: note,
				})
			}
		}
	}

	// Stable ordering for the rendered output.
	sort.Slice(rows, func(i, j int) bool { return rows[i].key < rows[j].key })

	if len(rows) == 0 {
		fmt.Println("preset-bump: no changes — preset already at the latest quarantined versions.")
		return
	}

	fmt.Println("preset-bump: proposed changes")
	for _, r := range rows {
		fmt.Printf("  %-30s %s → %s", r.key, formatPair(r.oldVer, r.newVer), formatSHA(r.oldSHA, r.newSHA))
		if r.note != "" {
			fmt.Printf("  %s", r.note)
		}
		fmt.Println()
	}

	if *dryRun {
		fmt.Println("(dry-run — pins.toml not modified)")
		return
	}

	data, err := encodePins(next)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: encode pins: %v\n", err)
		os.Exit(1)
	}
	path, err := filepath.Abs(*output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: resolve path: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: write %s: %v\n", path, err)
		os.Exit(1)
	}
	fmt.Printf("preset-bump: wrote %s\n", path)
}

// rawBundlesForBump exposes the un-pinned bundles for iteration. We
// re-use preset.All() rather than dipping into rawBundles directly
// (which is unexported); applyPins is a no-op when keys don't exist
// yet, so All() is equivalent for our purposes here.
func rawBundlesForBump() []preset.Bundle { return preset.All() }

func fetchSHA(ctx context.Context, url string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return "", fmt.Errorf("%s: %s", url, resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:]), nil
}

func formatPair(oldV, newV string) string {
	if oldV == newV {
		return oldV
	}
	if oldV == "" {
		return "(new) " + newV
	}
	return oldV + " → " + newV
}

func formatSHA(oldS, newS string) string {
	if oldS == newS {
		if oldS == "" {
			return ""
		}
		return "sha=" + oldS[:8] + "…"
	}
	if oldS == "" {
		return "sha=(new)" + newS[:8] + "…"
	}
	return "sha=" + oldS[:8] + "…→" + newS[:8] + "…"
}

// encodePins serializes with sorted keys for clean diffs. The
// BurntSushi/toml encoder doesn't guarantee key order across runs,
// so we build the output by hand.
func encodePins(ps preset.PinSet) ([]byte, error) {
	var b strings.Builder
	b.WriteString("# lfg version pin set.\n")
	b.WriteString("#\n")
	b.WriteString("# This file is maintained by .github/workflows/preset-bump.yml.\n")
	b.WriteString("# DO NOT edit by hand — the bumper regenerates the whole file.\n\n")
	fmt.Fprintf(&b, "bumped_at = %q\n\n", ps.BumpedAt.Format(time.RFC3339))

	keys := make([]string, 0, len(ps.Pins))
	for k := range ps.Pins {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		e := ps.Pins[k]
		fmt.Fprintf(&b, "[pins.%q]\n", k)
		if e.Version != "" {
			fmt.Fprintf(&b, "version = %q\n", e.Version)
		}
		if e.SHA256 != "" {
			fmt.Fprintf(&b, "sha256  = %q\n", e.SHA256)
		}
		b.WriteString("\n")
	}
	return []byte(b.String()), nil
}
