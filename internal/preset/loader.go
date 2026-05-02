package preset

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// presetFile is the on-disk schema. Top-level "bundles" array of Bundle.
type presetFile struct {
	Bundles []Bundle `toml:"bundles"`
}

// Load resolves a config spec (local path or http(s) URL) and returns
// the parsed bundles. Empty spec is treated as an error — callers
// should branch on the empty case before calling.
func Load(spec string) ([]Bundle, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, fmt.Errorf("empty preset spec")
	}
	if strings.HasPrefix(spec, "http://") || strings.HasPrefix(spec, "https://") {
		return LoadFromURL(spec)
	}
	return LoadFromPath(spec)
}

// LoadFromPath reads a TOML file from disk and parses it into bundles.
func LoadFromPath(path string) ([]Bundle, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return parse(data, path)
}

// LoadFromURL fetches a TOML file over HTTP(S) (10s hard timeout) and
// parses it. Caps the body at 1 MiB to keep a hostile URL from
// exhausting memory.
func LoadFromURL(url string) ([]Bundle, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: status %d", url, resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	return parse(data, url)
}

func parse(data []byte, source string) ([]Bundle, error) {
	var pf presetFile
	if err := toml.Unmarshal(data, &pf); err != nil {
		return nil, fmt.Errorf("parse %s: %w", source, err)
	}
	if len(pf.Bundles) == 0 {
		return nil, fmt.Errorf("parse %s: no [[bundles]] entries found", source)
	}
	return pf.Bundles, nil
}
