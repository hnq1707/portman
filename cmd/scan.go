package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/nay-kia/portman/internal/port"
)

var scanCount int

var scanCmd = &cobra.Command{
	Use:   "scan <start>-<end>",
	Short: "🔍 Scan for free ports in a range",
	Long:  "Scan a port range to find available (free) ports.\nUseful when you need to find a port for a new service.",
	Example: `  portman scan 8000-9000
  portman scan 3000-4000 --count 5
  portman scan 8080-8090`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		green := color.New(color.FgGreen, color.Bold)
		red := color.New(color.FgRed, color.Bold)
		cyan := color.New(color.FgCyan, color.Bold)
		dim := color.New(color.FgHiBlack)

		// Parse range
		parts := strings.Split(args[0], "-")
		if len(parts) != 2 {
			red.Printf("\n  ✗ Invalid range format. Use: start-end (e.g., 8000-9000)\n\n")
			return
		}

		start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err1 != nil || err2 != nil {
			red.Printf("\n  ✗ Invalid port numbers in range\n\n")
			return
		}

		if start < 1 || end > 65535 {
			red.Printf("\n  ✗ Port range must be between 1 and 65535\n\n")
			return
		}

		total := end - start + 1
		if total < 0 {
			total = -total
		}

		cyan.Printf("\n  🔍 Scanning %d ports (%d-%d)...\n\n", total, start, end)

		freePorts := port.FindFreePorts(start, end, scanCount)

		if len(freePorts) == 0 {
			red.Printf("  ✗ No free ports found in range %d-%d\n\n", start, end)
			return
		}

		green.Printf("  ✓ Found %d free port(s):\n\n", len(freePorts))

		for _, p := range freePorts {
			fmt.Printf("    ● %d\n", p)
		}

		fmt.Println()
		dim.Printf("  💡 Use any of these ports for your service.\n\n")
	},
}

func init() {
	scanCmd.Flags().IntVarP(&scanCount, "count", "c", 0, "Max number of free ports to find (0 = all)")
	rootCmd.AddCommand(scanCmd)
}
