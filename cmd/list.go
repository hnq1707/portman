package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/hnq1707/portman/internal/port"
	"github.com/hnq1707/portman/internal/ui"
)

var (
	listPort   int
	listJSON   bool
	listFilter string
	listDev    bool
	listRange  string
	listFree   string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "📡 List all listening ports",
	Long:  "Display all TCP/UDP ports currently in LISTENING state with process information.",
	Example: `  portman list
  portman list --port 8080
  portman list --json
  portman list --filter node
  portman list --free 8000-9000
  portman list --free 8000-9000 --count 5`,
	Run: func(cmd *cobra.Command, args []string) {
		if listFree != "" {
			runFreePortScan(cmd)
			return
		}
		runList(cmd, args)
	},
}

func runList(cmd *cobra.Command, args []string) {
	ports, err := fetchPorts(args)
	if err != nil {
		red := color.New(color.FgRed, color.Bold)
		red.Fprintf(cmd.ErrOrStderr(), "\n  ✗ Error: %v\n\n", err)
		return
	}

	if listJSON {
		if err := ui.RenderPortJSON(ports); err != nil {
			red := color.New(color.FgRed)
			red.Fprintf(cmd.ErrOrStderr(), "\n  ✗ JSON error: %v\n\n", err)
		}
		return
	}

	ui.RenderPortTable(ports)

	if listPort > 0 && len(ports) == 0 {
		dim := color.New(color.FgHiBlack)
		dim.Printf("  💡 Tip: port %d is free — no process is using it.\n\n", listPort)
	}

	// Summary line
	if len(ports) > 0 {
		dim := color.New(color.FgHiBlack)
		uniquePIDs := make(map[int]bool)
		for _, p := range ports {
			uniquePIDs[p.PID] = true
		}
		dim.Printf("  %d port(s), %d process(es)\n\n", len(ports), len(uniquePIDs))
	}
}

var freeCount int

func runFreePortScan(cmd *cobra.Command) {
	green := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	cyan := color.New(color.FgCyan, color.Bold)
	dim := color.New(color.FgHiBlack)

	// Parse range
	parts := strings.Split(listFree, "-")
	if len(parts) != 2 {
		red.Printf("\n  ✗ Invalid range format. Use: --free start-end (e.g., --free 8000-9000)\n\n")
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

	freePorts := port.FindFreePorts(start, end, freeCount)

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
}

func fetchPorts(args []string) ([]port.PortInfo, error) {
	var ports []port.PortInfo
	var err error

	if listPort > 0 {
		ports, err = port.FindByPort(listPort)
	} else {
		ports, err = port.ScanPorts()
	}
	if err != nil {
		return nil, err
	}

	// Positional arg filter: portman list 8080
	if listPort == 0 && len(args) > 0 {
		if p, e := strconv.Atoi(args[0]); e == nil {
			var filtered []port.PortInfo
			for _, entry := range ports {
				if entry.Port == p {
					filtered = append(filtered, entry)
				}
			}
			ports = filtered
		}
	}

	// Dev port range filter (3000-9999)
	if listDev {
		var filtered []port.PortInfo
		for _, entry := range ports {
			if entry.Port >= 3000 && entry.Port <= 9999 {
				filtered = append(filtered, entry)
			}
		}
		ports = filtered
	}

	// Custom range filter: --range 8000-9000
	if listRange != "" {
		rangeParts := strings.Split(listRange, "-")
		if len(rangeParts) == 2 {
			low, e1 := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			high, e2 := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if e1 == nil && e2 == nil {
				var filtered []port.PortInfo
				for _, entry := range ports {
					if entry.Port >= low && entry.Port <= high {
						filtered = append(filtered, entry)
					}
				}
				ports = filtered
			}
		}
	}

	// Process name filter
	if listFilter != "" {
		lower := strings.ToLower(listFilter)
		var filtered []port.PortInfo
		for _, entry := range ports {
			if strings.Contains(strings.ToLower(entry.ProcessName), lower) {
				filtered = append(filtered, entry)
			}
		}
		ports = filtered
	}

	return ports, nil
}

func init() {
	listCmd.Flags().IntVarP(&listPort, "port", "p", 0, "Filter by specific port number")
	listCmd.Flags().BoolVarP(&listJSON, "json", "j", false, "Output in JSON format")
	listCmd.Flags().StringVarP(&listFilter, "filter", "f", "", "Filter by process name")
	listCmd.Flags().BoolVarP(&listDev, "dev", "d", false, "Show only dev ports (3000-9999)")
	listCmd.Flags().StringVarP(&listRange, "range", "r", "", "Filter by port range (e.g. 8000-9000)")
	listCmd.Flags().StringVar(&listFree, "free", "", "Scan for free ports in a range (e.g. 8000-9000)")
	listCmd.Flags().IntVarP(&freeCount, "count", "c", 0, "Max free ports to find with --free (0 = all)")
	rootCmd.AddCommand(listCmd)
}
