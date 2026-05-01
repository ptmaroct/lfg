package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/ptmaroct/lfg/internal/detect"
	"github.com/ptmaroct/lfg/internal/installer"
	"github.com/ptmaroct/lfg/internal/preset"
)

var (
	applyYes bool
)

var applyCmd = &cobra.Command{
	Use:   "apply [bundle...]",
	Short: "Install bundles non-interactively",
	Long: `Headless equivalent of the TUI flow. Pass bundle IDs to install
(default: 'default'). Use --dry-run to print commands without running them.

Examples:
  lfg apply                       # install everything in the 'default' bundle
  lfg apply default ai-clis       # multiple bundles
  lfg apply --dry-run             # preview commands only`,
	RunE: runApply,
}

func init() {
	// --dry-run is inherited from the root persistent flag; only the
	// confirmation skip lives locally.
	applyCmd.Flags().BoolVarP(&applyYes, "yes", "y", false,
		"skip the confirmation prompt")
	rootCmd.AddCommand(applyCmd)
}

func runApply(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		args = []string{"default"}
	}
	wanted := map[string]bool{}
	for _, a := range args {
		wanted[a] = true
	}

	bundles := preset.All()
	if !dryRun {
		// Detect first so we can skip already-installed tools. Cheap on
		// a warm cache.
		results := detect.ProbeAll(bundles)
		bundles = detect.Apply(bundles, results)
	}

	// Build selection: every tool in every requested bundle, skipping
	// the ones detect already marked Installed.
	selected := map[string]bool{}
	picked := []string{}
	for _, b := range bundles {
		if !wanted[b.ID] {
			continue
		}
		picked = append(picked, b.ID)
		for _, t := range b.Tools {
			if t.Installed && !dryRun {
				continue
			}
			selected[b.ID+"/"+t.Name] = true
		}
	}
	if len(picked) == 0 {
		return fmt.Errorf("no matching bundles for %v (try `lfg --help`)", args)
	}

	plan := installer.Plan(bundles, selected)

	if dryRun {
		fmt.Printf("Plan for bundles %s — %d steps:\n", strings.Join(picked, ", "), len(plan))
		for i, step := range plan {
			label := fmt.Sprintf("[%s] %s", step.Backend, step.Tool.Name)
			if step.Bootstrap {
				label = fmt.Sprintf("[%s] (bootstrap)", step.Backend)
			}
			cmdline := installer.For(step.Backend).DryRun(step.Tool)
			if step.Bootstrap {
				cmdline = "(installs the installer itself)"
			}
			fmt.Printf("  %2d. %-32s  %s\n", i+1, label, cmdline)
		}
		return nil
	}

	if !applyYes {
		fmt.Printf("About to run %d install steps for: %s\n", len(plan), strings.Join(picked, ", "))
		fmt.Print("Continue? [y/N] ")
		var ans string
		fmt.Scanln(&ans)
		if !strings.EqualFold(strings.TrimSpace(ans), "y") {
			fmt.Println("aborted")
			return nil
		}
	}

	// Wire ctrl+c to cancel the run gracefully.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	out := make(chan installer.Line, 64)
	done := make(chan []installer.FailedStep)
	go func() {
		done <- installer.Run(ctx, plan, out)
		close(out)
	}()

	for line := range out {
		printLine(line)
	}
	failed := <-done

	if len(failed) > 0 {
		fmt.Fprintf(os.Stderr, "\n%d step(s) failed:\n", len(failed))
		for _, f := range failed {
			label := f.Step.Tool.Name
			if f.Step.Bootstrap {
				label = f.Step.Backend + " (bootstrap)"
			}
			fmt.Fprintf(os.Stderr, "  - %s: %v\n", label, f.Err)
		}
		return fmt.Errorf("apply finished with %d failure(s)", len(failed))
	}
	fmt.Println("\nDone.")
	return nil
}

func printLine(l installer.Line) {
	prefix := "  "
	switch l.Stream {
	case "meta":
		prefix = "» "
	case "stderr":
		prefix = "! "
	}
	fmt.Printf("%s%s | %s\n", prefix, l.Tool, l.Text)
}
