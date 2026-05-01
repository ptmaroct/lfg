// Package version exposes build-time metadata. The values are populated
// by goreleaser via -ldflags="-X github.com/ptmaroct/lfg/internal/version.Version=..."
// at release time. Defaults below cover `go run` / `go install` builds.
package version

// Version is the semver of the build, e.g. "v0.1.0". "dev" for unreleased.
var Version = "dev"

// Commit is the short git SHA. Empty unless set by ldflags.
var Commit = ""

// BuildDate is RFC3339, e.g. "2026-05-01T10:42:00Z". Empty unless set.
var BuildDate = ""

// Short returns a single-line string like "v0.1.0" or "dev (abc1234)".
func Short() string {
	if Commit == "" {
		return Version
	}
	return Version + " (" + Commit + ")"
}

// Full returns multi-line "lfg v0.1.0\ncommit abc1234\nbuilt 2026-..."
// for `lfg version --verbose`.
func Full() string {
	out := "lfg " + Version
	if Commit != "" {
		out += "\ncommit " + Commit
	}
	if BuildDate != "" {
		out += "\nbuilt " + BuildDate
	}
	return out
}
