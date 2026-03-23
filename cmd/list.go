package cmd

import (
	"fmt"
	"strconv"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/nay-kia/portman/internal/port"
	"github.com/nay-kia/portman/internal/ui"
)

var (
	listPort int
	listJSON bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "📡 List all listening ports",
	Long:  "Display all TCP/UDP ports currently in LISTENING state with process information.",
	Example: `  portman list
  portman list --port 8080
  portman list --json`,
	Run: func(cmd *cobra.Command, args []string) {
		var ports []port.PortInfo
		var err error

		if listPort > 0 {
			ports, err = port.FindByPort(listPort)
		} else {
			ports, err = port.ScanPorts()
		}

		if err != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(cmd.ErrOrStderr(), "\n  ✗ Error: %v\n\n", err)
			return
		}

		// Also support positional arg: portman list 8080
		if listPort == 0 && len(args) > 0 {
			if p, err := strconv.Atoi(args[0]); err == nil {
				var filtered []port.PortInfo
				for _, entry := range ports {
					if entry.Port == p {
						filtered = append(filtered, entry)
					}
				}
				ports = filtered
			}
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
	},
}

func init() {
	listCmd.Flags().IntVarP(&listPort, "port", "p", 0, "Filter by specific port number")
	listCmd.Flags().BoolVarP(&listJSON, "json", "j", false, "Output in JSON format")
	rootCmd.AddCommand(listCmd)
	_ = fmt.Sprintf("") // ensure fmt import
}
