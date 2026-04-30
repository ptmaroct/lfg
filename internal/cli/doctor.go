package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/anuj/lfg/internal/doctor"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose environment readiness for lfg",
	Long: `Runs a battery of checks (network, package managers, writable
config dir, supported shell) and prints results. Useful before running
` + "`lfg apply`" + ` on a new machine.

Exit code: 0 = all pass or warnings only; 1 = at least one failure.`,
	RunE: runDoctor,
}

func runDoctor(cmd *cobra.Command, args []string) error {
	results := doctor.Run()
	for _, r := range results {
		marker := badge(r.Status)
		name := doctor.PadName(r.Name, results)
		fmt.Printf("%s  %s  %s\n", marker, name, r.Detail)
		if r.Hint != "" && r.Status != doctor.Pass {
			fmt.Printf("        ↳ %s\n", r.Hint)
		}
	}
	fmt.Println()
	fmt.Println(doctor.SummaryLine(results))
	if doctor.HasFailures(results) {
		os.Exit(1)
	}
	return nil
}

func badge(s doctor.Status) string {
	switch s {
	case doctor.Pass:
		return "✓"
	case doctor.Warn:
		return "!"
	default:
		return "✗"
	}
}

func init() { rootCmd.AddCommand(doctorCmd) }
