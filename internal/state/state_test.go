package state

import (
	"os"
	"path/filepath"
	"testing"
)

// withTmpHome redirects state.json to a tmpdir for a single test.
func withTmpHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	// Mac falls back to $HOME/Library/Application Support — override that too.
	t.Setenv("HOME", dir)
	return filepath.Join(dir, "lfg")
}

func TestLoadMissingReturnsZero(t *testing.T) {
	withTmpHome(t)
	s, err := Load()
	if err != nil {
		t.Fatalf("Load on missing file: %v", err)
	}
	if s.Theme != "" || s.SchemaVer != 0 {
		t.Errorf("expected zero State, got %+v", s)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	withTmpHome(t)
	want := State{Theme: "dracula"}
	if err := Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Theme != want.Theme {
		t.Errorf("theme: got %q want %q", got.Theme, want.Theme)
	}
	if got.SchemaVer != CurrentSchemaVer {
		t.Errorf("schema_ver: got %d want %d", got.SchemaVer, CurrentSchemaVer)
	}
	if got.LastRun.IsZero() {
		t.Error("LastRun should be auto-stamped on Save")
	}
}

func TestSavePermissions(t *testing.T) {
	withTmpHome(t)
	if err := Save(State{Theme: "lfg"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	path, _ := ConfigPath()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("perm: got %v want 0600", info.Mode().Perm())
	}
}
