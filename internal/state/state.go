// Package state persists user preferences (theme, last-run timestamp, etc.)
// to ~/.config/lfg/state.json so settings survive across sessions.
//
// Schema is versioned via SchemaVer so future migrations stay sane. A
// missing or unreadable file is treated as "fresh machine" — Load returns
// a zero-value State and a nil error.
package state

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// CurrentSchemaVer is bumped when on-disk format changes.
const CurrentSchemaVer = 1

// State is what we persist. JSON-encoded at ConfigPath().
type State struct {
	Theme     string    `json:"theme,omitempty"`
	LastRun   time.Time `json:"last_run,omitempty"`
	SchemaVer int       `json:"schema_ver"`
}

// ConfigDir returns ~/.config/lfg (or platform equivalent). Creates the
// directory on first call.
func ConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "lfg")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

// ConfigPath returns the absolute path of state.json.
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "state.json"), nil
}

// Load reads state.json. Returns zero-value State + nil error when file
// is absent (fresh machine). Any other error (parse, permission) is
// surfaced as-is.
func Load() (State, error) {
	var s State
	path, err := ConfigPath()
	if err != nil {
		return s, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return s, nil
		}
		return s, err
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return s, err
	}
	return s, nil
}

// Save writes state.json with mode 0600. Always stamps SchemaVer +
// LastRun so callers don't have to remember.
func Save(s State) error {
	s.SchemaVer = CurrentSchemaVer
	if s.LastRun.IsZero() {
		s.LastRun = time.Now()
	}
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	// Write to temp + rename = atomic on POSIX, avoids torn writes.
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
