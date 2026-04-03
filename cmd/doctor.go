package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/nay-kia/portman/internal/doctor"
	"github.com/nay-kia/portman/internal/port"
)

var doctorFix bool

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "🏥 Diagnose port health issues",
	Long: `Run a comprehensive health check on all listening ports.

Detects:
  • Port conflicts (multiple processes on same port)
  • Port hogging (one process using too many ports)
  • Suspicious processes on unknown ports
  • Ephemeral port overuse
  • Privileged port usage

Use --fix to interactively resolve fixable issues.`,
	Example: `  portman doctor
  portman doctor --fix`,
	Run: func(cmd *cobra.Command, args []string) {
		cyan := color.New(color.FgCyan, color.Bold)
		red := color.New(color.FgRed, color.Bold)

		cyan.Println("\n  🏥 Running diagnostics...")

		report, err := doctor.RunDiagnostics()
		if err != nil {
			red.Printf("\n  ✗ %v\n\n", err)
			return
		}

		doctor.RenderReport(report)

		if doctorFix {
			runDoctorFix(report)
		}
	},
}

func runDoctorFix(report *doctor.Report) {
	green := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)
	dim := color.New(color.FgHiBlack)

	fixable := doctor.GetFixablePorts(report.Findings)
	if len(fixable) == 0 {
		green.Println("  ✓ No fixable issues found!")
		fmt.Println()
		return
	}

	yellow.Printf("  🔧 Found %d fixable port(s):\n\n", len(fixable))

	for _, f := range report.Findings {
		if !f.Fixable || f.Severity < doctor.SeverityWarning {
			continue
		}

		for _, p := range f.Ports {
			dim.Printf("    → :%d — %s\n", p, f.Title)

			yellow.Printf("    ? Fix this issue? [y/N]: ")
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))

			if answer == "y" || answer == "yes" {
				results := port.KillByPort(p)
				for _, r := range results {
					if r.Success {
						green.Printf("    ✓ Killed %s (PID %d) on port %d\n", r.ProcessName, r.PID, p)
					} else {
						red.Printf("    ✗ Failed: %s\n", r.Error)
					}
				}
			} else {
				dim.Println("    ↩ Skipped")
			}
			fmt.Println()
		}
	}
}

func init() {
	doctorCmd.Flags().BoolVar(&doctorFix, "fix", false, "Interactively fix detected issues")
	rootCmd.AddCommand(doctorCmd)
}
