// Package detect probes the local machine for installed CLI tools and
// their versions. Replaces the hardcoded `Tool.Installed` / `Tool.Version`
// values in the preset with real data so the picker shows actual state.
//
// Probing strategy per tool:
//  1. exec.LookPath(binary)         → if absent, Installed=false, done.
//  2. run "<bin> --version" (3s)    → exit 0 + regex match → Version.
//  3. fallback "<bin> version"      → covers `go`, etc.
//  4. binary present but no version → Installed=true, Version="".
//
// Results are returned as a map keyed "<bundleID>/<toolName>" matching the
// keys used by tree/confirm models. Concurrent probes via errgroup-style
// goroutine fan-out keep the welcome → tree transition snappy.
package detect

import (
	"bytes"
	"context"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/ptmaroct/lfg/internal/preset"
)

// Result is the per-tool probe outcome.
type Result struct {
	Installed bool
	Version   string // "" when probe couldn't extract one
	Path      string // exec.LookPath result; "" when not installed
}

// versionRe matches the first dotted version number on a line, e.g.
// "git version 2.42.0" → "2.42.0", "go version go1.26 darwin/arm64" → "1.26".
var versionRe = regexp.MustCompile(`\d+(?:\.\d+)+`)

// probeTimeout caps each subprocess. Some tools (e.g. mise) can be slow on
// cold start; 3s is comfortable without making the picker feel stuck.
const probeTimeout = 3 * time.Second

// Probe runs the detection sequence for a single tool. Synchronous.
func Probe(t preset.Tool) Result {
	bin := t.Binary
	if bin == "" {
		bin = t.Name
	}
	// Some preset names contain spaces / parens (e.g. "node (lts)"). Reject
	// anything non-binaryish so LookPath doesn't hit the filesystem with
	// nonsense.
	if strings.ContainsAny(bin, " ()") {
		// Hopeful fallback: first whitespace-delimited token.
		bin = strings.Fields(bin)[0]
	}

	path, err := exec.LookPath(bin)
	if err != nil {
		return Result{}
	}

	res := Result{Installed: true, Path: path}

	// Try --version, then `version` as fallback.
	for _, args := range [][]string{{"--version"}, {"version"}} {
		ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
		out, err := runVersionCmd(ctx, path, args...)
		cancel()
		if err != nil {
			continue
		}
		if v := versionRe.FindString(out); v != "" {
			res.Version = v
			return res
		}
	}
	return res
}

// runVersionCmd executes <path> <args...> with a hard timeout, returning
// combined stdout+stderr. Many tools print version on stderr (`bun -V`,
// some node versions), so we capture both.
func runVersionCmd(ctx context.Context, path string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, path, args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		// Even on non-zero exit, some tools print version first then error;
		// if buffer has something, treat it as success and let regex try.
		if buf.Len() > 0 {
			return buf.String(), nil
		}
		return "", err
	}
	return buf.String(), nil
}

// ProbeAll fans out Probe across all tools in all bundles. Concurrent —
// each probe is bounded by probeTimeout so worst-case wall time is one
// timeout regardless of how many tools are absent.
func ProbeAll(bundles []preset.Bundle) map[string]Result {
	type job struct {
		key string
		t   preset.Tool
	}
	var jobs []job
	for _, b := range bundles {
		for _, t := range b.Tools {
			jobs = append(jobs, job{key: b.ID + "/" + t.Name, t: t})
		}
	}

	results := make(map[string]Result, len(jobs))
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(len(jobs))
	for _, j := range jobs {
		go func(j job) {
			defer wg.Done()
			r := Probe(j.t)
			mu.Lock()
			results[j.key] = r
			mu.Unlock()
		}(j)
	}
	wg.Wait()
	return results
}

// ProbeStep is one streamed probe outcome. Emitted by ProbeAllStream
// once per tool so the TUI can show live progress instead of the
// single-shot ProbeAll waiting silently.
type ProbeStep struct {
	Key    string
	Tool   preset.Tool
	Result Result
}

// ProbeAllStream fans out probes like ProbeAll but emits each completion
// on `out`, closing the channel when all probes finish. Buffer `out`
// generously (≥ total tool count) or the goroutines will block. Use
// Apply on the bundles + the collected results map for the final shape.
func ProbeAllStream(bundles []preset.Bundle, out chan<- ProbeStep) {
	type job struct {
		key string
		t   preset.Tool
	}
	var jobs []job
	for _, b := range bundles {
		for _, t := range b.Tools {
			jobs = append(jobs, job{key: b.ID + "/" + t.Name, t: t})
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(jobs))
	for _, j := range jobs {
		go func(j job) {
			defer wg.Done()
			out <- ProbeStep{Key: j.key, Tool: j.t, Result: Probe(j.t)}
		}(j)
	}
	wg.Wait()
	close(out)
}

// harnessProbes lists the AI harnesses we can pass to `npx skills add -a`.
// Each entry maps the npx-skills agent name to the binary we LookPath for.
// Order is the priority order presented to the skills CLI when multiple
// harnesses are present — claude-code first since it's the lfg default.
var harnessProbes = []struct {
	Agent  string // -a flag value for `npx skills add`
	Binary string // binary to LookPath
}{
	{Agent: "claude-code", Binary: "claude"},
	{Agent: "codex", Binary: "codex"},
	{Agent: "opencode", Binary: "opencode"},
}

// DetectedHarnesses returns the npx-skills agent names corresponding to
// AI harness binaries currently on PATH. Empty result is possible when
// none of the supported harnesses are installed; callers should fall
// back to a default (the skills installer uses claude-code).
func DetectedHarnesses() []string {
	var found []string
	for _, h := range harnessProbes {
		if _, err := exec.LookPath(h.Binary); err == nil {
			found = append(found, h.Agent)
		}
	}
	return found
}

// Apply overlays detect results onto bundle data, returning a new slice
// with Tool.Installed / Tool.Version reflecting reality. Pure — does not
// mutate input. Tools missing from results retain their preset values
// (useful when tests inject a partial map).
func Apply(bundles []preset.Bundle, results map[string]Result) []preset.Bundle {
	out := make([]preset.Bundle, len(bundles))
	for i, b := range bundles {
		nb := b
		nb.Tools = make([]preset.Tool, len(b.Tools))
		for j, t := range b.Tools {
			nt := t
			if r, ok := results[b.ID+"/"+t.Name]; ok {
				nt.Installed = r.Installed
				nt.Version = r.Version
			}
			nb.Tools[j] = nt
		}
		out[i] = nb
	}
	return out
}
