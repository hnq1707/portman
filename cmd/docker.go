package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/nay-kia/portman/internal/docker"
)

var dockerCmd = &cobra.Command{
	Use:     "docker",
	Short:   "🐳 List Docker container port bindings",
	Long:    "Show port mappings from all running Docker containers.",
	Aliases: []string{"dk"},
	Run: func(cmd *cobra.Command, args []string) {
		cyan := color.New(color.FgCyan, color.Bold)
		red := color.New(color.FgRed, color.Bold)
		yellow := color.New(color.FgYellow)
		green := color.New(color.FgGreen)
		dim := color.New(color.FgHiBlack)
		white := color.New(color.FgWhite)
		sepColor := color.New(color.FgHiBlack)

		containers, err := docker.ListContainerPorts()
		if err != nil {
			red.Printf("\n  ✗ %v\n\n", err)
			return
		}

		if len(containers) == 0 {
			yellow.Println("\n  ⚠ No running Docker containers found.")
			dim.Println("  Start some containers with: docker-compose up")
			fmt.Println()
			return
		}

		// Count total port bindings
		totalPorts := 0
		for _, c := range containers {
			totalPorts += len(c.Ports)
		}

		cyan.Printf("\n  🐳 Docker containers: %d running, %d port(s)\n\n", len(containers), totalPorts)

		// Header
		headerColor := color.New(color.FgHiWhite, color.Bold)
		widths := []int{20, 25, 10, 10, 8, 12}
		headers := []string{"CONTAINER", "IMAGE", "HOST PORT", "CONT PORT", "PROTO", "STATUS"}

		fmt.Print("  ")
		for i, h := range headers {
			headerColor.Printf("%-*s", widths[i], h)
			if i < len(headers)-1 {
				sepColor.Print("│ ")
			}
		}
		fmt.Println()

		// Separator
		fmt.Print("  ")
		for i, w := range widths {
			sepColor.Print(strings.Repeat("─", w))
			if i < len(widths)-1 {
				sepColor.Print("┼─")
			}
		}
		fmt.Println()

		// Rows
		for _, c := range containers {
			if len(c.Ports) == 0 {
				fmt.Print("  ")
				green.Printf("%-*s", widths[0], truncate(c.Name, widths[0]-1))
				sepColor.Print("│ ")
				white.Printf("%-*s", widths[1], truncate(c.Image, widths[1]-1))
				sepColor.Print("│ ")
				dim.Printf("%-*s", widths[2], "-")
				sepColor.Print("│ ")
				dim.Printf("%-*s", widths[3], "-")
				sepColor.Print("│ ")
				dim.Printf("%-*s", widths[4], "-")
				sepColor.Print("│ ")
				statusColor(c.Status).Printf("%-*s", widths[5], truncate(c.Status, widths[5]-1))
				fmt.Println()
				continue
			}

			for i, p := range c.Ports {
				fmt.Print("  ")
				if i == 0 {
					green.Printf("%-*s", widths[0], truncate(c.Name, widths[0]-1))
				} else {
					fmt.Printf("%-*s", widths[0], "")
				}
				sepColor.Print("│ ")

				if i == 0 {
					white.Printf("%-*s", widths[1], truncate(c.Image, widths[1]-1))
				} else {
					fmt.Printf("%-*s", widths[1], "")
				}
				sepColor.Print("│ ")

				// Host port
				if p.HostPort != "-" {
					yellow := color.New(color.FgYellow, color.Bold)
					yellow.Printf("%-*s", widths[2], p.HostPort)
				} else {
					dim.Printf("%-*s", widths[2], "-")
				}
				sepColor.Print("│ ")

				white.Printf("%-*s", widths[3], p.ContainerPort)
				sepColor.Print("│ ")

				dim.Printf("%-*s", widths[4], p.Proto)
				sepColor.Print("│ ")

				if i == 0 {
					statusColor(c.Status).Printf("%-*s", widths[5], truncate(c.Status, widths[5]-1))
				}
				fmt.Println()
			}
		}

		fmt.Println()
	},
}

func statusColor(status string) *color.Color {
	s := strings.ToLower(status)
	if strings.Contains(s, "up") {
		return color.New(color.FgGreen)
	}
	return color.New(color.FgRed)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-2] + ".."
}

func init() {
	rootCmd.AddCommand(dockerCmd)
	_ = os.Stdout // ensure os import
}
