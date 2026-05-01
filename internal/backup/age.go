package backup

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"filippo.io/age"

	"github.com/ptmaroct/lfg/internal/state"
)

// keyFilename is where the master age identity is stored. Sibling to
// state.json under the lfg config dir.
const keyFilename = "key.txt"

// LoadOrCreateIdentity returns the age identity used to encrypt
// backups, generating a fresh one on first call. The file is written
// with mode 0600. Caller must back it up out-of-band — losing the key
// means losing every encrypted backup.
//
// Returns the identity, the absolute path of the key file, and a bool
// indicating whether the key was newly created (true → caller should
// surface a "back this up!" warning to the user).
func LoadOrCreateIdentity() (*age.X25519Identity, string, bool, error) {
	dir, err := state.ConfigDir()
	if err != nil {
		return nil, "", false, err
	}
	path := filepath.Join(dir, keyFilename)

	if b, err := os.ReadFile(path); err == nil {
		// First non-comment, non-blank line is the identity.
		line := firstSecretLine(string(b))
		if line == "" {
			return nil, path, false, fmt.Errorf("%s: no identity line found", path)
		}
		id, err := age.ParseX25519Identity(line)
		if err != nil {
			return nil, path, false, fmt.Errorf("%s: %w", path, err)
		}
		return id, path, false, nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return nil, path, false, err
	}

	// Generate.
	id, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, path, false, err
	}
	content := fmt.Sprintf("# created by lfg %s\n# public key: %s\n%s\n",
		"v0.1", id.Recipient(), id)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return nil, path, false, err
	}
	return id, path, true, nil
}

// firstSecretLine returns the first line of `s` that isn't blank or a
// `#` comment. Mirrors how age's own `age-keygen` writes identity files.
func firstSecretLine(s string) string {
	for _, ln := range strings.Split(s, "\n") {
		ln = strings.TrimSpace(ln)
		if ln == "" || strings.HasPrefix(ln, "#") {
			continue
		}
		return ln
	}
	return ""
}

// encryptStream wraps a writer with age encryption. Caller must close
// the returned WriteCloser before the underlying writer (the encryption
// finalizer writes the trailer there).
func encryptStream(dst io.Writer, recipient age.Recipient) (io.WriteCloser, error) {
	return age.Encrypt(dst, recipient)
}
