package preset

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

const sample = `
[[bundles]]
id = "minimal"
name = "minimal"
default = true

  [[bundles.tools]]
  name = "git"
  source = "brew"
`

const sampleWithAliases = sample + `
[[aliases]]
name = "ll"
command = "ls -la"
group = "shell"
default = true
`

func TestLoadFromPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "preset.toml")
	if err := os.WriteFile(path, []byte(sample), 0o600); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded.Bundles) != 1 || loaded.Bundles[0].ID != "minimal" {
		t.Fatalf("bad bundles: %+v", loaded.Bundles)
	}
	if len(loaded.Bundles[0].Tools) != 1 || loaded.Bundles[0].Tools[0].Name != "git" {
		t.Fatalf("bad tools: %+v", loaded.Bundles[0].Tools)
	}
	if len(loaded.Aliases) != 0 {
		t.Fatalf("expected no aliases, got %+v", loaded.Aliases)
	}
}

func TestLoadFromURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(sample))
	}))
	defer srv.Close()

	loaded, err := Load(srv.URL)
	if err != nil {
		t.Fatalf("Load url: %v", err)
	}
	if len(loaded.Bundles) != 1 || loaded.Bundles[0].ID != "minimal" {
		t.Fatalf("bad bundles: %+v", loaded.Bundles)
	}
}

func TestLoadEmpty(t *testing.T) {
	if _, err := Load(""); err == nil {
		t.Fatal("expected error on empty spec")
	}
}

func TestLoadBadPath(t *testing.T) {
	if _, err := Load("/nonexistent/path.toml"); err == nil {
		t.Fatal("expected error on missing file")
	}
}

func TestLoadMalformed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.toml")
	if err := os.WriteFile(path, []byte("this is not = valid toml ["), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestLoadEmptyBundles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.toml")
	if err := os.WriteFile(path, []byte("# nothing here"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected error on no [[bundles]]")
	}
}

func TestLoadWithAliases(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "aliases.toml")
	if err := os.WriteFile(path, []byte(sampleWithAliases), 0o600); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded.Aliases) != 1 || loaded.Aliases[0].Name != "ll" {
		t.Fatalf("bad aliases: %+v", loaded.Aliases)
	}
	if loaded.Aliases[0].Command != "ls -la" {
		t.Fatalf("bad alias command: %q", loaded.Aliases[0].Command)
	}
}
