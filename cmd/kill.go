package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/nay-kia/portman/internal/port"
)

var forceKill bool

var killCmd = &cobra.Command{
	Use:   "kill <port> [port2] [port3...]",
	Short: "💀 Kill process(es) on specified port(s)",
	Long:  "Find and terminate processes listening on the specified port(s).",
	Example: `  portman kill 8080
  portman kill 3000 8080 5432
  portman kill 8080 --force`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		green := color.New(color.FgGreen, color.Bold)
		red := color.New(color.FgRed, color.Bold)
		yellow := color.New(color.FgYellow)
		dim := color.New(color.FgHiBlack)

		fmt.Println()

		for _, arg := range args {
			portNum, err := strconv.Atoi(arg)
			if err != nil {
				red.Printf("  ✗ Invalid port number: %s\n", arg)
				continue
			}

			// Find what's on this port first
			entries, err := port.FindByPort(portNum)
			if err != nil {
				red.Printf("  ✗ Error scanning port %d: %v\n", portNum, err)
				continue
			}

			if len(entries) == 0 {
				yellow.Printf("  ⚠ Port %d — no process found\n", portNum)
				continue
			}

			// Show what we'll kill
			for _, entry := range entries {
				dim.Printf("  → Port %d: %s (PID %d)\n", portNum, entry.ProcessName, entry.PID)
			}

			// Confirm unless --force
			if !forceKill {
				yellow.Printf("  ? Kill process(es) on port %d? [y/N]: ", portNum)
				reader := bufio.NewReader(os.Stdin)
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(strings.ToLower(answer))
				if answer != "y" && answer != "yes" {
					dim.Printf("  ↩ Skipped port %d\n", portNum)
					continue
				}
			}

			// Kill it
			results := port.KillByPort(portNum)
			for _, r := range results {
				if r.Success {
					green.Printf("  ✓ Killed %s (PID %d) on port %d\n", r.ProcessName, r.PID, portNum)
				} else {
					red.Printf("  ✗ Failed to kill PID %d on port %d: %s\n", r.PID, portNum, r.Error)
				}
			}
		}

		fmt.Println()
	},
}

func init() {
	killCmd.Flags().BoolVarP(&forceKill, "force", "f", false, "Skip confirmation prompt")
	rootCmd.AddCommand(killCmd)
}
