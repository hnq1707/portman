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
	"github.com/nay-kia/portman/internal/utils"
)

var (
	forceKill   bool
	killAllName string
)

var killCmd = &cobra.Command{
	Use:   "kill <port> [port2] [port3...]",
	Short: "💀 Kill process(es) on specified port(s)",
	Long:  "Find and terminate processes listening on the specified port(s).\nUse --all to kill all ports of a specific process.",
	Example: `  portman kill 8080
  portman kill 3000 8080 5432
  portman kill 8080 --force
  portman kill --all java
  portman kill --all node -f`,
	Run: func(cmd *cobra.Command, args []string) {
		if killAllName != "" {
			runKillAll(cmd)
			return
		}
		if len(args) == 0 {
			red := color.New(color.FgRed, color.Bold)
			red.Println("\n  ✗ Please specify port(s) to kill, or use --all <process>")
			cmd.Usage()
			return
		}
		runKillPorts(cmd, args)
	},
}

func runKillPorts(cmd *cobra.Command, args []string) {
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

		entries, err := port.FindByPort(portNum)
		if err != nil {
			red.Printf("  ✗ Error scanning port %d: %v\n", portNum, err)
			continue
		}

		if len(entries) == 0 {
			yellow.Printf("  ⚠ Port %d — no process found\n", portNum)
			continue
		}

		for _, entry := range entries {
			dim.Printf("  → Port %d: %s (PID %d)\n", portNum, entry.ProcessName, entry.PID)
		}

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

		results := port.KillByPort(portNum)
		for _, r := range results {
			if r.Success {
				green.Printf("  ✓ Killed %s (PID %d) on port %d\n", r.ProcessName, r.PID, portNum)
			} else {
				red.Printf("  ✗ Failed to kill PID %d on port %d: %s\n", r.PID, portNum, r.Error)
				if strings.Contains(r.Error, "Access Denied") && !utils.IsAdmin() {
					yellow.Println("     💡 Tip: Try running PortMan in an Administrator terminal.")
				}
			}
		}
	}

	fmt.Println()
}

func runKillAll(cmd *cobra.Command) {
	green := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	yellow := color.New(color.FgYellow)
	dim := color.New(color.FgHiBlack)
	cyan := color.New(color.FgCyan, color.Bold)

	fmt.Println()

	lower := strings.ToLower(killAllName)

	allPorts, err := port.ScanPorts()
	if err != nil {
		red.Printf("  ✗ Error scanning ports: %v\n\n", err)
		return
	}

	// Find matching processes
	var matched []port.PortInfo
	for _, p := range allPorts {
		if strings.Contains(strings.ToLower(p.ProcessName), lower) {
			matched = append(matched, p)
		}
	}

	if len(matched) == 0 {
		yellow.Printf("  ⚠ No ports found for process matching '%s'\n\n", killAllName)
		return
	}

	cyan.Printf("  🎯 Found %d port(s) for '%s':\n\n", len(matched), killAllName)

	for _, entry := range matched {
		dim.Printf("    → :%d %s (PID %d)\n", entry.Port, entry.ProcessName, entry.PID)
	}
	fmt.Println()

	if !forceKill {
		yellow.Printf("  ? Kill ALL %d port(s) of '%s'? [y/N]: ", len(matched), killAllName)
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			dim.Println("  ↩ Cancelled")
			fmt.Println()
			return
		}
	}

	// Kill all matched — deduplicate by PID
	killed := make(map[int]bool)
	for _, entry := range matched {
		if killed[entry.PID] {
			continue
		}

		results := port.KillByPort(entry.Port)
		for _, r := range results {
			killed[r.PID] = true
			if r.Success {
				green.Printf("  ✓ Killed %s (PID %d) on port %d\n", r.ProcessName, r.PID, entry.Port)
			} else {
				red.Printf("  ✗ Failed: PID %d — %s\n", r.PID, r.Error)
			}
		}
	}

	fmt.Println()
}

func init() {
	killCmd.Flags().BoolVarP(&forceKill, "force", "f", false, "Skip confirmation prompt")
	killCmd.Flags().StringVarP(&killAllName, "all", "a", "", "Kill all ports of a process by name")
	rootCmd.AddCommand(killCmd)
}
