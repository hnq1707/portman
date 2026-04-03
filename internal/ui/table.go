package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"

	"github.com/hnq1707/portman/internal/port"
)

// wellKnownPorts is now centralized in port.WellKnownPorts

// RenderPortTable renders port info as a pretty table.
func RenderPortTable(ports []port.PortInfo) {
	if len(ports) == 0 {
		yellow := color.New(color.FgYellow)
		yellow.Println("\n  ⚠ No listening ports found.")
		return
	}

	cyan := color.New(color.FgCyan, color.Bold)
	cyan.Printf("\n  📡 Found %d listening port(s)\n\n", len(ports))

	// Column widths
	headers := []string{"PROTO", "PORT", "LOCAL ADDR", "PID", "PROCESS", "NOTE"}
	widths := []int{7, 7, 22, 8, 20, 18}

	// Compute max widths from data
	for _, p := range ports {
		note := port.GetPortLabel(p.Port)
		cols := []string{
			p.Proto,
			strconv.Itoa(p.Port),
			p.LocalAddr,
			strconv.Itoa(p.PID),
			p.ProcessName,
			note,
		}
		for i, c := range cols {
			if len(c)+2 > widths[i] {
				widths[i] = len(c) + 2
			}
		}
	}

	// Header
	headerColor := color.New(color.FgHiWhite, color.Bold)
	sepColor := color.New(color.FgHiBlack)

	fmt.Print("  ")
	for i, h := range headers {
		headerColor.Printf("%-*s", widths[i], h)
		if i < len(headers)-1 {
			sepColor.Print("│ ")
		}
	}
	fmt.Println()

	// Separator line
	fmt.Print("  ")
	for i, w := range widths {
		sepColor.Print(strings.Repeat("─", w))
		if i < len(widths)-1 {
			sepColor.Print("┼─")
		}
	}
	fmt.Println()

	// Rows
	cyanC := color.New(color.FgCyan)
	yellowC := color.New(color.FgYellow, color.Bold)
	whiteC := color.New(color.FgWhite)
	dimC := color.New(color.FgHiBlack)
	greenC := color.New(color.FgGreen)
	hiYellowC := color.New(color.FgHiYellow)

	for _, p := range ports {
		note := port.GetPortLabel(p.Port)
		isHighlight := note != ""
		if isHighlight {
			note = "⭐ " + note
		}

		fmt.Print("  ")

		// PROTO
		cyanC.Printf("%-*s", widths[0], p.Proto)
		sepColor.Print("│ ")

		// PORT
		if isHighlight {
			yellowC.Printf("%-*s", widths[1], strconv.Itoa(p.Port))
		} else {
			whiteC.Printf("%-*s", widths[1], strconv.Itoa(p.Port))
		}
		sepColor.Print("│ ")

		// LOCAL ADDR
		whiteC.Printf("%-*s", widths[2], p.LocalAddr)
		sepColor.Print("│ ")

		// PID
		dimC.Printf("%-*s", widths[3], strconv.Itoa(p.PID))
		sepColor.Print("│ ")

		// PROCESS
		if isHighlight {
			greenC.Printf("%-*s", widths[4], p.ProcessName)
		} else {
			whiteC.Printf("%-*s", widths[4], p.ProcessName)
		}
		sepColor.Print("│ ")

		// NOTE
		if isHighlight {
			hiYellowC.Printf("%-*s", widths[5], note)
		} else {
			fmt.Printf("%-*s", widths[5], note)
		}

		fmt.Println()
	}

	fmt.Println()
}

// RenderPortJSON renders port info as JSON.
func RenderPortJSON(ports []port.PortInfo) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(ports)
}
