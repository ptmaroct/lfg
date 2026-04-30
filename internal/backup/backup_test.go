package backup

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// fakeHome creates a sandbox dir with a few representative dotfiles +
// SSH keys, points homeDir at it for the test, and returns the dir.
func fakeHome(t *testing.T) string {
	t.Helper()
	d := t.TempDir()
	mustWrite(t, d, ".zshrc", "alias ll=ls\n")
	mustWrite(t, d, ".gitconfig", "[user]\n  email=test@example.com\n")
	mustWrite(t, d, ".ssh/config", "Host *\n  ServerAliveInterval 60\n")
	mustWrite(t, d, ".ssh/id_ed25519", "PRIVATE-KEY-NEVER-EXPORT")
	mustWrite(t, d, ".ssh/id_ed25519.pub", "ssh-ed25519 AAAA...")

	prev := homeDir
	homeDir = func() string { return d }
	t.Cleanup(func() { homeDir = prev })
	return d
}

func mustWrite(t *testing.T, root, name, content string) {
	t.Helper()
	p := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestPack_PlainGzipRoundTrip(t *testing.T) {
	fakeHome(t)
	out := t.TempDir()
	r, err := Pack(Options{
		OutDir:   out,
		Encrypt:  false,
		Hostname: "tester",
		Now:      time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("Pack: %v", err)
	}

	want := filepath.Join(out, "lfg-backup-tester-2026-05-01.tar.gz")
	if r.Path != want {
		t.Errorf("path: got %q want %q", r.Path, want)
	}
	if r.Files == 0 {
		t.Error("expected at least one file in the archive")
	}

	// Read it back.
	names := readGzTarNames(t, r.Path)
	mustContain(t, names, ".zshrc")
	mustContain(t, names, ".gitconfig")
	mustContain(t, names, ".ssh/config")
	mustContain(t, names, ".ssh/id_ed25519.pub")
	mustNotContain(t, names, ".ssh/id_ed25519")
}

func TestPack_ExcludesPrivateKeysByDefault(t *testing.T) {
	fakeHome(t)
	out := t.TempDir()
	r, err := Pack(Options{OutDir: out, Encrypt: false, Hostname: "h", Now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	if r.Excluded == 0 {
		t.Error("expected at least one private-key exclusion")
	}
	for _, n := range readGzTarNames(t, r.Path) {
		if n == ".ssh/id_ed25519" {
			t.Errorf("private key leaked into archive: %s", n)
		}
	}
}

func TestPack_EncryptProducesAgeFile(t *testing.T) {
	fakeHome(t)
	// Force the age key to live inside the sandbox. On Mac
	// os.UserConfigDir reads $HOME/Library/Application Support; on Linux
	// it honors XDG_CONFIG_HOME. Set both so the test is portable.
	cfg := t.TempDir()
	t.Setenv("HOME", cfg)
	t.Setenv("XDG_CONFIG_HOME", cfg)

	out := t.TempDir()
	r, err := Pack(Options{OutDir: out, Encrypt: true, Hostname: "h",
		Now: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("Pack: %v", err)
	}
	if filepath.Ext(r.Path) != ".age" {
		t.Errorf("expected .age suffix, got %s", r.Path)
	}
	if !r.NewKey {
		t.Error("expected NewKey=true on first encrypt")
	}
	if _, err := os.Stat(r.KeyPath); err != nil {
		t.Errorf("key file should exist: %v", err)
	}
	// Subsequent encrypt reuses the key.
	r2, err := Pack(Options{OutDir: out, Encrypt: true, Hostname: "h2",
		Now: time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	if r2.NewKey {
		t.Error("second encrypt should reuse existing key")
	}
}

// readGzTarNames opens a .tar.gz and returns the list of entry names.
func readGzTarNames(t *testing.T, path string) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	var names []string
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		names = append(names, h.Name)
	}
	return names
}

func mustContain(t *testing.T, names []string, want string) {
	t.Helper()
	for _, n := range names {
		if n == want {
			return
		}
	}
	t.Errorf("archive missing %q; entries: %v", want, names)
}

func mustNotContain(t *testing.T, names []string, banned string) {
	t.Helper()
	for _, n := range names {
		if n == banned {
			t.Errorf("archive must not contain %q", banned)
		}
	}
}
