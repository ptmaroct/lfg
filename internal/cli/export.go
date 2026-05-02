package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/ptmaroct/lfg/internal/preset"
)

var exportOut string

var exportCmd = &cobra.Command{
	Use:   "export [path]",
	Short: "Save the current preset to a TOML file",
	Long: `Export bundles + tools as a TOML file matching the schema accepted by
--config. Captures the install recipe only — no dotfile content, no
secrets — so the result is safe to share or check into a dotfiles repo.

Default output: ~/lfg-preset-YYYY-MM-DD.toml`,
	RunE: runExport,
}

func init() {
	exportCmd.Flags().StringVarP(&exportOut, "out", "o", "",
		"output path (default: ~/lfg-preset-<date>.toml)")
	rootCmd.AddCommand(exportCmd)
}

func runExport(cmd *cobra.Command, args []string) error {
	bundles, err := loadPreset()
	if err != nil {
		return err
	}

	out := exportOut
	if len(args) > 0 {
		out = args[0]
	}
	if out == "" {
		home, _ := os.UserHomeDir()
		out = filepath.Join(home, fmt.Sprintf("lfg-preset-%s.toml",
			time.Now().Format("2006-01-02")))
	}

	if dryRun {
		fmt.Printf("(dry-run) would write %s with %d bundles\n", out, len(bundles))
		return nil
	}

	if err := preset.Save(out, bundles); err != nil {
		return err
	}
	fmt.Printf("✓ saved preset to %s\n", out)
	fmt.Printf("  re-load with:  lfg --config %s\n", out)
	return nil
}
