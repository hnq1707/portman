package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/nay-kia/portman/internal/port"
	"github.com/nay-kia/portman/internal/ui"
)

var (
	listPort   int
	listJSON   bool
	listWatch  bool
	listFilter string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "📡 List all listening ports",
	Long:  "Display all TCP/UDP ports currently in LISTENING state with process information.",
	Example: `  portman list
  portman list --port 8080
  portman list --json
  portman list --watch
  portman list --filter node`,
	Run: func(cmd *cobra.Command, args []string) {
		if listWatch {
			runWatch(args)
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

func runWatch(args []string) {
	cyan := color.New(color.FgCyan, color.Bold)
	dim := color.New(color.FgHiBlack)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	render := func() {
		// Clear screen
		fmt.Print("\033[2J\033[H")

		cyan.Println("  📡 PORTMAN WATCH MODE")
		dim.Printf("  Refreshing every 2s • Press Ctrl+C to stop\n")

		ports, err := fetchPorts(args)
		if err != nil {
			red := color.New(color.FgRed)
			red.Printf("\n  ✗ Error: %v\n", err)
			return
		}

		ui.RenderPortTable(ports)

		if len(ports) > 0 {
			uniquePIDs := make(map[int]bool)
			for _, p := range ports {
				uniquePIDs[p.PID] = true
			}
			dim.Printf("  %d port(s), %d process(es) • %s\n",
				len(ports), len(uniquePIDs), time.Now().Format("15:04:05"))
		}
	}

	// First render
	render()

	for {
		select {
		case <-ticker.C:
			render()
		case <-sigCh:
			fmt.Print("\033[2J\033[H")
			yellow := color.New(color.FgYellow)
			yellow.Println("\n  ⏹ Watch mode stopped.\n")
			return
		}
	}
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
	listCmd.Flags().BoolVarP(&listWatch, "watch", "w", false, "Auto-refresh every 2 seconds")
	listCmd.Flags().StringVarP(&listFilter, "filter", "f", "", "Filter by process name")
	rootCmd.AddCommand(listCmd)
	_ = fmt.Sprintf("") // ensure fmt import
}
