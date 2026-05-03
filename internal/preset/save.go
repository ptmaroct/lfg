package preset

import (
	"bytes"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Save writes bundles + aliases back to a TOML file matching the
// loader's schema. Used by `lfg export` and the in-TUI "save preset"
// flow so users can capture their machine's tool set as a portable
// file and re-load it elsewhere with `lfg --config <file>`.
func Save(path string, bundles []Bundle, aliases []Alias) error {
	if path == "" {
		return fmt.Errorf("empty output path")
	}
	pf := presetFile{Bundles: bundles, Aliases: aliases}
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	enc.Indent = "  "
	if err := enc.Encode(pf); err != nil {
		return fmt.Errorf("encode toml: %w", err)
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}
