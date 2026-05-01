// Package doctor runs a battery of environment-readiness checks. It's
// the diagnostic counterpart of `lfg apply` — surfaces missing
// dependencies, unsupported shells, write-protected config dirs, and
// the like before an install run.
//
// Each Check is a self-contained func returning a Result. Checks are
// pure-ish (may stat the filesystem, may spawn short subprocesses) but
// must not mutate global state.
package doctor

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/ptmaroct/lfg/internal/state"
)

// Status enumerates the three Check outcomes.
type Status int

const (
	Pass Status = iota
	Warn
	Fail
)

func (s Status) String() string {
	switch s {
	case Pass:
		return "PASS"
	case Warn:
		return "WARN"
	default:
		return "FAIL"
	}
}

// Result is one check's outcome.
type Result struct {
	Name   string
	Status Status
	Detail string // short user-facing line; empty for trivial pass
	Hint   string // remediation; only set on Warn/Fail
}

// Check is the unit interface — each registered check returns a Result.
type Check func() Result

// All returns the set of checks for the host platform. Order is the
// order they're presented in the output.
func All() []Check {
	checks := []Check{
		checkNetwork,
		checkConfigDir,
		checkShell,
		checkCurl,
	}
	if runtime.GOOS == "darwin" {
		checks = append(checks, checkBrew, checkXcode)
	} else {
		checks = append(checks, checkApt)
	}
	checks = append(checks, checkMise, checkNode)
	return checks
}

// Run executes every check serially and returns the results.
func Run() []Result {
	cs := All()
	out := make([]Result, len(cs))
	for i, c := range cs {
		out[i] = c()
	}
	return out
}

// ---------- individual checks ----------

func checkNetwork() Result {
	r := Result{Name: "network"}
	client := &http.Client{Timeout: 4 * time.Second}
	resp, err := client.Get("https://1.1.1.1")
	if err != nil {
		r.Status = Fail
		r.Detail = "no internet: " + err.Error()
		r.Hint = "check Wi-Fi / DNS; lfg can't fetch presets or install packages offline"
		return r
	}
	resp.Body.Close()
	r.Status = Pass
	return r
}

func checkConfigDir() Result {
	r := Result{Name: "config dir"}
	dir, err := state.ConfigDir()
	if err != nil {
		r.Status = Fail
		r.Detail = err.Error()
		return r
	}
	// Confirm write access via a probe file.
	probe := filepath.Join(dir, ".write-probe")
	if err := os.WriteFile(probe, []byte("ok"), 0o600); err != nil {
		r.Status = Fail
		r.Detail = "not writable: " + err.Error()
		r.Hint = "fix permissions on " + dir
		return r
	}
	os.Remove(probe)
	r.Status = Pass
	r.Detail = dir
	return r
}

func checkShell() Result {
	r := Result{Name: "shell"}
	sh := os.Getenv("SHELL")
	if sh == "" {
		r.Status = Warn
		r.Detail = "$SHELL is empty"
		r.Hint = "run `chsh -s /bin/zsh` (or your preferred shell)"
		return r
	}
	base := filepath.Base(sh)
	switch base {
	case "zsh", "bash":
		r.Status = Pass
		r.Detail = sh
	default:
		r.Status = Warn
		r.Detail = sh + " (untested)"
		r.Hint = "lfg's preset rc snippets target zsh/bash; YMMV elsewhere"
	}
	return r
}

func checkCurl() Result { return checkBinary("curl", true) }
func checkBrew() Result { return checkBinary("brew", false) }
func checkApt() Result  { return checkBinary("apt-get", true) }
func checkMise() Result { return checkBinary("mise", false) }
func checkNode() Result { return checkBinary("node", false) }

// checkXcode verifies the macOS Command Line Tools are present.
// Without them brew (and most compilers) won't install. `xcode-select
// -p` exits 0 with a path when installed, 2 when not.
func checkXcode() Result {
	r := Result{Name: "xcode tools"}
	cmd := exec.Command("xcode-select", "-p")
	if err := cmd.Run(); err != nil {
		r.Status = Fail
		r.Detail = "Command Line Tools not installed"
		r.Hint = "run `xcode-select --install` and accept the dialog"
		return r
	}
	r.Status = Pass
	return r
}

// checkBinary returns Pass when the binary is on PATH, otherwise Warn
// (or Fail when `required` is true). The ternary saves a stack of
// boilerplate per check.
func checkBinary(name string, required bool) Result {
	r := Result{Name: name}
	path, err := exec.LookPath(name)
	if err == nil {
		r.Status = Pass
		r.Detail = path
		return r
	}
	if required {
		r.Status = Fail
		r.Detail = "not on PATH"
		r.Hint = "install " + name + " — required by lfg"
		return r
	}
	r.Status = Warn
	r.Detail = "not on PATH"
	r.Hint = "lfg can install this for you via `lfg apply`"
	return r
}

// SummaryLine is a concise one-line summary of a Run result for
// ”X passed · Y warnings · Z failures” style reporting.
func SummaryLine(rs []Result) string {
	var p, w, f int
	for _, r := range rs {
		switch r.Status {
		case Pass:
			p++
		case Warn:
			w++
		case Fail:
			f++
		}
	}
	return fmt.Sprintf("%d passed · %d warnings · %d failed", p, w, f)
}

// HasFailures returns true iff any Result is Fail.
func HasFailures(rs []Result) bool {
	for _, r := range rs {
		if r.Status == Fail {
			return true
		}
	}
	return false
}

// PadName left-pads a check name to the longest in the set so the
// terminal output stays column-aligned.
func PadName(name string, all []Result) string {
	max := 0
	for _, r := range all {
		if n := len(r.Name); n > max {
			max = n
		}
	}
	if pad := max - len(name); pad > 0 {
		return name + strings.Repeat(" ", pad)
	}
	return name
}
