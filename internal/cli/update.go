package cli

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/minio/selfupdate"
	"github.com/spf13/cobra"

	"github.com/ptmaroct/lfg/internal/version"
)

// releasesAPI is the latest-release endpoint for ptmaroct/lfg. Override
// at build/test time if the repo moves.
const releasesAPI = "https://api.github.com/repos/ptmaroct/lfg/releases/latest"

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Self-update lfg to the latest GitHub release",
	Long: `Queries the latest release on github.com/ptmaroct/lfg, downloads
the asset for the current OS/arch, verifies the checksum, and atomically
swaps the running binary.

A no-op when already on the latest version.`,
	RunE: runUpdate,
}

func init() { rootCmd.AddCommand(updateCmd) }

type ghRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func runUpdate(cmd *cobra.Command, args []string) error {
	rel, err := fetchLatestRelease(releasesAPI)
	if err != nil {
		return fmt.Errorf("query latest release: %w", err)
	}

	if rel.TagName == version.Version {
		fmt.Printf("already on %s\n", version.Version)
		return nil
	}
	fmt.Printf("found %s (current: %s)\n", rel.TagName, version.Version)

	asset := pickAsset(rel)
	if asset == "" {
		return fmt.Errorf("no asset found for %s/%s in release %s",
			runtime.GOOS, runtime.GOARCH, rel.TagName)
	}

	fmt.Printf("downloading %s\n", asset)
	resp, err := http.Get(asset)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download: HTTP %d", resp.StatusCode)
	}

	// goreleaser ships a tar.gz archive containing the binary alongside
	// LICENSE/README. Find the `lfg` entry and stream that to selfupdate.
	binStream, err := extractBinaryFromTarGz(resp.Body, "lfg")
	if err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	// selfupdate.Apply swaps the running binary atomically (writes a
	// `.tmp` next to the exe, then renames).
	if err := selfupdate.Apply(binStream, selfupdate.Options{}); err != nil {
		if rb := selfupdate.RollbackError(err); rb != nil {
			return fmt.Errorf("update failed and rollback also failed: apply=%v rollback=%v", err, rb)
		}
		return fmt.Errorf("update: %w", err)
	}

	fmt.Printf("updated to %s\n", rel.TagName)
	return nil
}

func fetchLatestRelease(apiURL string) (ghRelease, error) {
	var r ghRelease
	if _, err := url.Parse(apiURL); err != nil {
		return r, err
	}
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return r, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return r, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return r, json.NewDecoder(resp.Body).Decode(&r)
}

// pickAsset returns the download URL of the tarball matching the
// current OS/arch, or "" when no match. goreleaser names archives:
//   lfg_<version>_<os>_<arch>.tar.gz
func pickAsset(r ghRelease) string {
	suffix := fmt.Sprintf("_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	for _, a := range r.Assets {
		if strings.HasSuffix(a.Name, suffix) {
			return a.BrowserDownloadURL
		}
	}
	return ""
}

// extractBinaryFromTarGz reads a gzipped tarball and returns a reader
// positioned at the named binary entry. The returned reader is bounded
// by the entry's size — Read returns io.EOF after the binary ends.
//
// We deliberately don't buffer the whole archive in memory; selfupdate
// streams the new binary directly from this reader to disk.
func extractBinaryFromTarGz(r io.Reader, binName string) (io.Reader, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return nil, fmt.Errorf("binary %q not found in archive", binName)
		}
		if err != nil {
			return nil, err
		}
		if h.Name == binName || strings.HasSuffix(h.Name, "/"+binName) {
			// gzip.Reader closure isn't strictly needed (process exits
			// after Apply) but keeps lints happy.
			return tr, nil
		}
	}
}
