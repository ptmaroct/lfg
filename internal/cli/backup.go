package cli

import (
	"fmt"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/ptmaroct/lfg/internal/backup"
)

var (
	backupEncrypt        bool
	backupOutDir         string
	backupIncludeSSHKeys bool
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Snapshot this machine's config to a single file",
	Long: `lfg backup gathers shell configs, dotfiles, package lists, and
optional secrets into a single tar (or tar.age, encrypted) file you can
copy to any new machine.

The age identity is stored at ~/.config/lfg/key.txt — back this file up
out-of-band, otherwise encrypted backups cannot be restored.`,
	RunE: runBackup,
}

func runBackup(cmd *cobra.Command, args []string) error {
	opts := backup.Options{
		OutDir:         backupOutDir,
		Encrypt:        backupEncrypt,
		IncludeSSHKeys: backupIncludeSSHKeys,
	}
	if dryRun {
		r, err := backup.Plan(opts)
		if err != nil {
			return err
		}
		fmt.Printf("(dry-run) would write %s\n", r.Path)
		fmt.Printf("  %d files · %s · %d sources missing · %d private keys filtered\n",
			r.Files, humanize.Bytes(uint64(r.Bytes)), r.Skipped, r.Excluded)
		return nil
	}
	r, err := backup.Pack(opts)
	if err != nil {
		return err
	}
	fmt.Printf("✓ wrote %s\n", r.Path)
	fmt.Printf("  %d files · %s · %d skipped · %d excluded (private keys)\n",
		r.Files, humanize.Bytes(uint64(r.Bytes)), r.Skipped, r.Excluded)
	if r.Encrypted {
		if r.NewKey {
			fmt.Printf("\n⚠  generated new age key at %s\n", r.KeyPath)
			fmt.Println("   BACK THIS FILE UP. Without it the backup is unrecoverable.")
		} else {
			fmt.Printf("  (encrypted with key at %s)\n", r.KeyPath)
		}
	}
	return nil
}

func init() {
	backupCmd.Flags().BoolVar(&backupEncrypt, "encrypt", true,
		"encrypt output with age (uses ~/.config/lfg/key.txt)")
	backupCmd.Flags().StringVarP(&backupOutDir, "out", "o", ".",
		"directory to write the backup file into")
	backupCmd.Flags().BoolVar(&backupIncludeSSHKeys, "include-ssh-keys", false,
		"DANGER: include private SSH keys (only when --encrypt is set)")
	rootCmd.AddCommand(backupCmd)
}
