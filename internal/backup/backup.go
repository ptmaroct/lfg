// Package backup snapshots a developer machine's configuration into a
// single restorable file. v0.1 captures shell rc files, common dotfiles,
// curated `~/.config/*` subdirs, AI-tool settings, and `~/.ssh/`
// (excluding private keys by default).
//
// Output is a tarball, optionally encrypted with age. The age identity
// lives at `~/.config/lfg/key.txt` — the user is responsible for backing
// it up out-of-band, otherwise encrypted snapshots are unrecoverable.
package backup

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Options controls a single Pack run.
type Options struct {
	// OutDir is where the backup file is written. "" → current dir.
	OutDir string
	// Encrypt wraps the tar in age. Output suffix: .tar.age (vs .tar.gz).
	Encrypt bool
	// IncludeSSHKeys lifts the default block on `~/.ssh/id_*` private
	// keys. Only honored when Encrypt=true (private keys must never
	// land in plaintext archives).
	IncludeSSHKeys bool
	// Hostname is embedded in the filename. "" → os.Hostname().
	Hostname string
	// Now overrides the timestamp. Zero → time.Now().
	Now time.Time
	// Sources overrides the default whitelist (tests use this).
	Sources []Source
}

// Result describes a successful Pack outcome.
type Result struct {
	Path        string // absolute path of the produced archive
	Files       int    // count of files included
	Bytes       int64  // total uncompressed bytes
	Encrypted   bool
	KeyPath     string // age identity location (set when Encrypted=true)
	NewKey      bool   // true when the identity was generated this run
	Skipped     int    // sources missing from disk (silent skips)
	Excluded    int    // private-key entries blocked by policy
}

// Plan walks the configured sources without opening any output file
// and returns the Result we *would* produce. Used by `lfg backup
// --dry-run` and the doctor / preview surfaces. Pure read-only.
func Plan(opts Options) (Result, error) {
	r := Result{Encrypted: opts.Encrypt}
	r.Path = outputPath(opts)
	sources := opts.Sources
	if sources == nil {
		sources = defaultSources()
	}
	for _, s := range sources {
		info, err := os.Lstat(s.Path)
		if err != nil {
			r.Skipped++
			continue
		}
		count, bytes, excluded := 0, int64(0), 0
		walk := func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() {
				return nil
			}
			if !(opts.IncludeSSHKeys && opts.Encrypt) && isPrivateSSHKey(path) {
				excluded++
				return nil
			}
			ii, err := os.Lstat(path)
			if err != nil {
				return nil
			}
			count++
			bytes += ii.Size()
			return nil
		}
		if info.IsDir() {
			_ = filepath.WalkDir(s.Path, walk)
		} else {
			_ = walk(s.Path, fs.FileInfoToDirEntry(info), nil)
		}
		r.Files += count
		r.Bytes += bytes
		r.Excluded += excluded
	}
	return r, nil
}

// outputPath builds the destination filename without creating any file.
// Shared by Plan and Pack so dry-run output matches the real run.
func outputPath(opts Options) string {
	out := opts.OutDir
	if out == "" {
		out = "."
	}
	host := opts.Hostname
	if host == "" {
		h, err := os.Hostname()
		if err != nil {
			h = "host"
		}
		host = sanitize(h)
	}
	stamp := opts.Now
	if stamp.IsZero() {
		stamp = time.Now()
	}
	ext := ".tar.gz"
	if opts.Encrypt {
		ext = ".tar.age"
	}
	return filepath.Join(out, fmt.Sprintf("lfg-backup-%s-%s%s", host, stamp.Format("2006-01-02"), ext))
}

// Pack captures the configured sources to a single file. Returns the
// Result with the absolute output path. Failure to read an individual
// source is logged via the Result counters, not fatal.
func Pack(opts Options) (Result, error) {
	var r Result
	r.Encrypted = opts.Encrypt

	if opts.OutDir == "" {
		opts.OutDir = "."
	}
	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return r, fmt.Errorf("mkdir %s: %w", opts.OutDir, err)
	}
	r.Path = outputPath(opts)

	f, err := os.OpenFile(r.Path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return r, fmt.Errorf("create %s: %w", r.Path, err)
	}
	defer f.Close()

	// Pipeline: tar.Writer → (gzip OR age) → file.
	var sink io.WriteCloser = nopCloser{f}
	if opts.Encrypt {
		id, keyPath, isNew, err := LoadOrCreateIdentity()
		if err != nil {
			return r, fmt.Errorf("age key: %w", err)
		}
		r.KeyPath = keyPath
		r.NewKey = isNew
		w, err := encryptStream(f, id.Recipient())
		if err != nil {
			return r, fmt.Errorf("age encrypt: %w", err)
		}
		sink = w
	} else {
		gz := gzip.NewWriter(f)
		sink = gz
	}
	tw := tar.NewWriter(sink)

	sources := opts.Sources
	if sources == nil {
		sources = defaultSources()
	}

	for _, s := range sources {
		n, bytes, excluded, err := writeSource(tw, s, opts.IncludeSSHKeys && opts.Encrypt)
		if errors.Is(err, fs.ErrNotExist) {
			r.Skipped++
			continue
		}
		if err != nil {
			return r, fmt.Errorf("source %s: %w", s.Path, err)
		}
		r.Files += n
		r.Bytes += bytes
		r.Excluded += excluded
	}

	if err := tw.Close(); err != nil {
		return r, fmt.Errorf("tar close: %w", err)
	}
	if err := sink.Close(); err != nil {
		return r, fmt.Errorf("close sink: %w", err)
	}
	return r, nil
}

// writeSource adds one Source to the tar. Recurses through directories.
// Returns (filesAdded, bytesAdded, excludedCount, err).
//
// excludedCount tracks files skipped due to private-key policy.
func writeSource(tw *tar.Writer, s Source, allowSSHKeys bool) (int, int64, int, error) {
	info, err := os.Lstat(s.Path)
	if err != nil {
		return 0, 0, 0, err
	}
	files, bytes, excluded := 0, int64(0), 0
	walk := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		// Block private SSH keys unless explicitly opted-in. Match by
		// name pattern (id_rsa, id_ed25519, etc.).
		if !allowSSHKeys && isPrivateSSHKey(path) {
			excluded++
			return nil
		}
		// Resolve archive name relative to source root.
		rel, err := filepath.Rel(s.Path, path)
		if err != nil {
			return err
		}
		archiveName := s.ArchiveName
		if rel != "." {
			archiveName = filepath.Join(s.ArchiveName, rel)
		}
		n, err := writeFile(tw, path, archiveName)
		if err != nil {
			return err
		}
		files++
		bytes += n
		return nil
	}
	if info.IsDir() {
		if err := filepath.WalkDir(s.Path, walk); err != nil {
			return files, bytes, excluded, err
		}
	} else {
		if err := walk(s.Path, fs.FileInfoToDirEntry(info), nil); err != nil {
			return files, bytes, excluded, err
		}
	}
	return files, bytes, excluded, nil
}

// writeFile copies a single file into the tar. Returns bytes written.
func writeFile(tw *tar.Writer, path, archiveName string) (int64, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return 0, err
	}
	hdr, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return 0, err
	}
	hdr.Name = archiveName
	if err := tw.WriteHeader(hdr); err != nil {
		return 0, err
	}
	if !info.Mode().IsRegular() {
		return 0, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return io.Copy(tw, f)
}

// isPrivateSSHKey returns true for files matching ~/.ssh/id_* (the
// canonical OpenSSH private-key naming) without a `.pub` suffix.
func isPrivateSSHKey(path string) bool {
	dir := filepath.Base(filepath.Dir(path))
	if dir != ".ssh" {
		return false
	}
	base := filepath.Base(path)
	if !strings.HasPrefix(base, "id_") {
		return false
	}
	if strings.HasSuffix(base, ".pub") {
		return false
	}
	return true
}

// sanitize strips characters that would be awkward in filenames.
func sanitize(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '-', r == '_':
			return r
		}
		return '-'
	}, s)
}

// nopCloser wraps a Writer with a no-op Close so the pipeline always
// terminates in a Closer.
type nopCloser struct{ w io.Writer }

func (n nopCloser) Write(p []byte) (int, error) { return n.w.Write(p) }
func (n nopCloser) Close() error                 { return nil }
