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

func TestLoadFromPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "preset.toml")
	if err := os.WriteFile(path, []byte(sample), 0o600); err != nil {
		t.Fatal(err)
	}
	bundles, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(bundles) != 1 || bundles[0].ID != "minimal" {
		t.Fatalf("bad bundles: %+v", bundles)
	}
	if len(bundles[0].Tools) != 1 || bundles[0].Tools[0].Name != "git" {
		t.Fatalf("bad tools: %+v", bundles[0].Tools)
	}
}

func TestLoadFromURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(sample))
	}))
	defer srv.Close()

	bundles, err := Load(srv.URL)
	if err != nil {
		t.Fatalf("Load url: %v", err)
	}
	if len(bundles) != 1 || bundles[0].ID != "minimal" {
		t.Fatalf("bad bundles: %+v", bundles)
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
