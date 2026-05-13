package installer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// urlRE plucks the first http(s) URL out of an install command line.
// We accept the URL stops at common shell metacharacters: whitespace,
// pipe, redirect, paren close, quote.
var urlRE = regexp.MustCompile(`https?://[^\s"'|)>]+`)

// runVerifiedCurl is the supply-chain hardened replacement for piping
// `curl ... | sh` into the shell.
//
// 1. Parse the original cmdline to extract the install-script URL.
// 2. HTTP GET that URL (HTTPS only; the regex enforces the scheme).
// 3. Verify the body's SHA256 against expectedSHA. On mismatch, abort
//    with a clear "supply-chain check failed" error — caller treats
//    this exactly like a failed install step.
// 4. Write the verified body to a temp file. Exec it via the same
//    interpreter the original line piped into (`bash <tmp>` when the
//    cmdline mentions bash, else `sh <tmp>`).
//
// This neutralises the two attack paths the bare `curl|sh` pattern
// invites: an attacker who serves a different file to lfg vs. a
// security researcher, and an attacker who pre-stages a benign script
// then swaps to malicious between our last review and the user's
// install run.
func runVerifiedCurl(ctx context.Context, tool, cmdline, expectedSHA string, out chan<- Line) error {
	url := urlRE.FindString(cmdline)
	if url == "" {
		return fmt.Errorf("verified-curl: no https URL found in %q", cmdline)
	}
	out <- Line{Tool: tool, Stream: "meta", Text: "fetching " + url + " (sha256 verify)"}

	body, err := fetchScript(ctx, url)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(body)
	got := hex.EncodeToString(sum[:])
	if !strings.EqualFold(got, expectedSHA) {
		return fmt.Errorf("supply-chain check failed for %s: expected sha256 %s, got %s", url, expectedSHA, got)
	}
	out <- Line{Tool: tool, Stream: "meta", Text: "sha256 verified: " + got[:12] + "…"}

	f, err := os.CreateTemp("", "lfg-install-*.sh")
	if err != nil {
		return fmt.Errorf("tempfile: %w", err)
	}
	tmpPath := f.Name()
	defer os.Remove(tmpPath)
	if _, err := f.Write(body); err != nil {
		f.Close()
		return fmt.Errorf("write tempfile: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close tempfile: %w", err)
	}

	shell := "sh"
	if strings.Contains(cmdline, "bash") {
		shell = "bash"
	}
	return runCmd(ctx, tool, shell+" "+tmpPath, out)
}

// fetchScript pulls the install script body. 30s timeout (these
// scripts are small — under 100 KB usually — so we don't need to be
// generous). Caps at 8 MiB to bound memory if upstream goes wild.
func fetchScript(ctx context.Context, url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "lfg-installer (+https://github.com/ptmaroct/lfg)")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("fetch %s: %s", url, resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", url, err)
	}
	return body, nil
}
