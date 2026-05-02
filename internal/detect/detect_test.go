package detect

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ptmaroct/lfg/internal/preset"
)

// TestProbe_AbsentBinary confirms a definitely-missing binary returns a
// zero Result. Uses an unrealistic name so the test doesn't accidentally
// match a real tool on someone's machine.
func TestProbe_AbsentBinary(t *testing.T) {
	r := Probe(preset.Tool{Name: "this-binary-does-not-exist-xyz123"})
	if r.Installed {
		t.Errorf("expected Installed=false, got %+v", r)
	}
}

// TestProbe_RealTool exercises the real probe path against a binary that
// is virtually guaranteed to be present on any dev machine running tests
// (`go` itself). Skipped if go isn't found — defensive for unusual envs.
func TestProbe_RealTool(t *testing.T) {
	r := Probe(preset.Tool{Name: "go"})
	if !r.Installed {
		t.Skip("go binary not found on PATH; skipping real-tool probe")
	}
	if r.Version == "" {
		t.Errorf("expected non-empty version for go, got Result=%+v", r)
	}
}

func TestApply_OverlaysResults(t *testing.T) {
	bundles := []preset.Bundle{
		{
			ID: "x",
			Tools: []preset.Tool{
				{Name: "a", Installed: false, Version: ""},
				{Name: "b", Installed: true, Version: "1.0"}, // pre-existing
			},
		},
	}
	results := map[string]Result{
		"x/a": {Installed: true, Version: "9.9.9"},
		// "x/b" intentionally omitted → preset values must survive.
	}
	got := Apply(bundles, results)
	if !got[0].Tools[0].Installed || got[0].Tools[0].Version != "9.9.9" {
		t.Errorf("a: not overlaid: %+v", got[0].Tools[0])
	}
	if !got[0].Tools[1].Installed || got[0].Tools[1].Version != "1.0" {
		t.Errorf("b: should retain preset values: %+v", got[0].Tools[1])
	}
	// Confirm input slice was not mutated.
	if bundles[0].Tools[0].Installed {
		t.Error("Apply mutated input slice")
	}
}

func TestApply_EmptyResults(t *testing.T) {
	bundles := []preset.Bundle{{ID: "x", Tools: []preset.Tool{{Name: "a", Version: "1"}}}}
	got := Apply(bundles, nil)
	if got[0].Tools[0].Version != "1" {
		t.Errorf("empty results should leave preset untouched: %+v", got[0].Tools[0])
	}
}

func TestProbeSkill(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	tool := preset.Tool{Name: "agent-browser", Source: "skills"}

	// Not installed yet.
	if r := Probe(tool); r.Installed {
		t.Fatalf("expected skill missing, got %+v", r)
	}

	// Cross-harness directory.
	dir := filepath.Join(home, ".agents", "skills", "agent-browser")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	r := Probe(tool)
	if !r.Installed || r.Path != dir {
		t.Fatalf("expected installed at %s, got %+v", dir, r)
	}

	// Falls back to ~/.claude/skills/<name> when ~/.agents not present.
	home2 := t.TempDir()
	t.Setenv("HOME", home2)
	claude := filepath.Join(home2, ".claude", "skills", "agent-browser")
	if err := os.MkdirAll(claude, 0o700); err != nil {
		t.Fatal(err)
	}
	r = Probe(tool)
	if !r.Installed || r.Path != claude {
		t.Fatalf("expected claude install at %s, got %+v", claude, r)
	}
}

// TestVersionRegex sanity-checks the version extractor against known
// outputs from real tools. Adding cases here is cheap — every time we
// find a tool whose `--version` parsing breaks, lock the case in.
func TestVersionRegex(t *testing.T) {
	cases := map[string]string{
		"git version 2.42.0":                        "2.42.0",
		"go version go1.26 darwin/arm64":            "1.26",
		"jq-1.7.1":                                  "1.7.1",
		"bat 0.24.0":                                "0.24.0",
		"v22.0.0":                                   "22.0.0",
		"ripgrep 14.1.1\nfeatures: ...":             "14.1.1",
		"":                                          "",
		"some text without numbers":                 "",
	}
	for in, want := range cases {
		got := versionRe.FindString(in)
		if got != want {
			t.Errorf("regex(%q) = %q, want %q", in, got, want)
		}
	}
}
