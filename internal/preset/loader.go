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

// presetFile is the on-disk schema. Top-level "bundles" array of Bundle
// plus an optional top-level "aliases" array of Alias for shell aliases.
type presetFile struct {
	Bundles []Bundle `toml:"bundles"`
	Aliases []Alias  `toml:"aliases,omitempty"`
}

// Loaded is the parsed config. Bundles is required, Aliases is optional
// (empty when the config omits [[aliases]] — opted-out by omission).
type Loaded struct {
	Bundles []Bundle
	Aliases []Alias
}

// Load resolves a config spec (local path or http(s) URL) and returns
// the parsed bundles + aliases. Empty spec is treated as an error —
// callers should branch on the empty case before calling.
func Load(spec string) (Loaded, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return Loaded{}, fmt.Errorf("empty preset spec")
	}
	if strings.HasPrefix(spec, "http://") || strings.HasPrefix(spec, "https://") {
		return LoadFromURL(spec)
	}
	return LoadFromPath(spec)
}

// LoadFromPath reads a TOML file from disk.
func LoadFromPath(path string) (Loaded, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Loaded{}, fmt.Errorf("read %s: %w", path, err)
	}
	return parse(data, path)
}

// LoadFromURL fetches a TOML file over HTTP(S) (10s hard timeout) and
// parses it. Caps the body at 1 MiB.
func LoadFromURL(url string) (Loaded, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Loaded{}, fmt.Errorf("build request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Loaded{}, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Loaded{}, fmt.Errorf("fetch %s: status %d", url, resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Loaded{}, fmt.Errorf("read body: %w", err)
	}
	return parse(data, url)
}

func parse(data []byte, source string) (Loaded, error) {
	var pf presetFile
	if err := toml.Unmarshal(data, &pf); err != nil {
		return Loaded{}, fmt.Errorf("parse %s: %w", source, err)
	}
	if len(pf.Bundles) == 0 {
		return Loaded{}, fmt.Errorf("parse %s: no [[bundles]] entries found", source)
	}
	return Loaded{Bundles: pf.Bundles, Aliases: pf.Aliases}, nil
}
