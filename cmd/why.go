package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/nay-kia/portman/internal/port"
)

// portLabels is now centralized in port.WellKnownPorts

var whyCmd = &cobra.Command{
	Use:   "why <port>",
	Short: "🔍 Deep investigate what's using a port",
	Long: `Investigate a port with full process details and interactive actions.

Shows: process info, command line, process tree, memory, related ports,
and config file references. Then lets you take actions interactively.`,
	Example: `  portman why 3000
  portman why 8080
  portman why 5432`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		red := color.New(color.FgRed, color.Bold)
		cyan := color.New(color.FgCyan, color.Bold)

		portNum, err := strconv.Atoi(args[0])
		if err != nil {
			red.Printf("\n  ✗ Invalid port: %s\n\n", args[0])
			return
		}

		cyan.Printf("\n  🔍 Investigating port %d...\n", portNum)

		detail, err := port.InvestigatePort(portNum)
		if err != nil {
			red.Printf("\n  ✗ %v\n\n", err)
			return
		}

		renderInvestigation(detail)
		runInteractive(detail)
	},
}

func renderInvestigation(d *port.ProcessDetail) {
	cyan := color.New(color.FgCyan, color.Bold)
	white := color.New(color.FgWhite, color.Bold)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	dim := color.New(color.FgHiBlack)
	magenta := color.New(color.FgMagenta)

	sep := "  ──────────────────────────────────────────────────"

	fmt.Println()
	dim.Println(sep)

	// Header
	label := port.GetPortLabel(d.Port)
	if label != "" {
		label = " (" + label + ")"
	}
	cyan.Printf("  🔍 PORT %d%s — %s\n", d.Port, label, d.ProcessName)
	dim.Println(sep)

	// ── Process Info ──
	fmt.Println()
	white.Println("  📋 Process Info")
	printInfoField(dim, "Process", fmt.Sprintf("%s (PID %d)", d.ProcessName, d.PID))

	if d.CommandLine != "" {
		cmdDisplay := d.CommandLine
		if len(cmdDisplay) > 60 {
			cmdDisplay = cmdDisplay[:57] + "..."
		}
		printInfoField(dim, "Command", cmdDisplay)
	}
	if d.ExecutablePath != "" {
		printInfoField(dim, "Exe Path", d.ExecutablePath)
	}
	if d.WorkingDir != "" {
		printInfoField(dim, "Directory", d.WorkingDir)
	}
	if d.User != "" {
		printInfoField(dim, "User", d.User)
	}
	if d.MemoryMB > 0 {
		memStr := fmt.Sprintf("%.1f MB", d.MemoryMB)
		if d.MemoryMB >= 1024 {
			memStr = fmt.Sprintf("%.2f GB", d.MemoryMB/1024)
		}
		printInfoField(dim, "Memory", memStr)
	}
	if !d.StartTime.IsZero() {
		elapsed := time.Since(d.StartTime)
		printInfoField(dim, "Started", fmt.Sprintf("%s (%s)",
			port.FormatDuration(elapsed),
			d.StartTime.Format("15:04:05")))
	}

	// ── Process Tree ──
	if len(d.ProcessTree) > 0 {
		fmt.Println()
		white.Println("  🌳 Process Tree")
		for _, node := range d.ProcessTree {
			indent := strings.Repeat("   ", node.Depth)
			connector := "└─ "
			if node.IsCurrent {
				yellow.Printf("     %s%s%s (PID %d) ← THIS\n",
					indent, connector, node.Name, node.PID)
			} else {
				dim.Printf("     %s%s%s (%d)\n",
					indent, connector, node.Name, node.PID)
			}
		}
	}

	// ── Related Ports ──
	if len(d.RelatedPorts) > 0 {
		fmt.Println()
		white.Println("  🔗 Related Ports (same process)")
		for _, p := range d.RelatedPorts {
			note := ""
			if l := port.GetPortLabel(p.Port); l != "" {
				note = " " + l
			}
			green.Printf("     • :%d%s\n", p.Port, note)
		}
	}

	// ── Config Files ──
	if len(d.ConfigFiles) > 0 {
		fmt.Println()
		white.Println("  📁 Config References")
		for _, cf := range d.ConfigFiles {
			baseName := filepath.Base(cf.FilePath)
			line := cf.Line
			if len(line) > 45 {
				line = line[:42] + "..."
			}
			magenta.Printf("     • %s:%d  ", baseName, cf.LineNum)
			dim.Printf("%s\n", line)
		}
	}

	fmt.Println()
}

func printInfoField(dim *color.Color, label, value string) {
	labelFmt := fmt.Sprintf("%-12s", label)
	dim.Printf("     %s ", labelFmt)
	fmt.Printf("%s\n", value)
}

func runInteractive(detail *port.ProcessDetail) {
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	yellow := color.New(color.FgYellow)
	dim := color.New(color.FgHiBlack)

	dim.Println("  ──────────────────────────────────────────────────")
	cyan.Printf("  Actions: ")
	fmt.Print("[k] Kill  ")
	if detail.WorkingDir != "" {
		fmt.Print("[o] Open dir  ")
	}
	fmt.Println("[q] Quit")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		cyan.Printf("  → ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(strings.ToLower(scanner.Text()))

		switch input {
		case "k":
			yellow.Printf("\n  ⚠ Kill %s (PID %d) on port %d? [y/n]: ",
				detail.ProcessName, detail.PID, detail.Port)
			if scanner.Scan() {
				confirm := strings.TrimSpace(strings.ToLower(scanner.Text()))
				if confirm == "y" || confirm == "yes" {
					results := port.KillByPort(detail.Port)
					if len(results) > 0 && results[0].Success {
						green.Printf("\n  ✓ Killed %s (PID %d) on port %d\n\n",
							detail.ProcessName, detail.PID, detail.Port)
					} else {
						errMsg := "unknown error"
						if len(results) > 0 {
							errMsg = results[0].Error
						}
						red.Printf("\n  ✗ Failed: %s\n\n", errMsg)
					}
					return
				}
				dim.Println("  ↩ Cancelled")
			}

		case "o":
			if detail.WorkingDir != "" {
				if err := openDirectory(detail.WorkingDir); err != nil {
					red.Printf("  ✗ Failed to open: %v\n", err)
				} else {
					green.Printf("  ✓ Opened %s\n", detail.WorkingDir)
				}
			} else {
				yellow.Println("  ⚠ Working directory unknown")
			}

		case "q", "":
			fmt.Println()
			return

		default:
			dim.Println("  Unknown command. Use [k]ill, [o]pen, or [q]uit")
		}
	}
}

func openDirectory(dir string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("explorer", dir).Start()
	case "darwin":
		return exec.Command("open", dir).Start()
	default:
		return exec.Command("xdg-open", dir).Start()
	}
}

func init() {
	rootCmd.AddCommand(whyCmd)
}
