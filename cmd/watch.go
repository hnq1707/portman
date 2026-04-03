package cmd

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/nay-kia/portman/internal/watch"
)

var watchInterval int

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "👁 Real-time port activity timeline",
	Long: `Monitor port changes in real-time with a diff-based timeline view.

Shows a compact port table and a live timeline log of all port changes:
  🟢 NEW  — a new port started listening
  🔴 GONE — a port stopped listening
  🔄 SWAP — a port changed process

Perfect for monitoring docker-compose up, service restarts, and debugging.`,
	Aliases: []string{"w"},
	Example: `  portman watch
  portman watch --interval 5`,
	Run: func(cmd *cobra.Command, args []string) {
		interval := time.Duration(watchInterval) * time.Second
		m := watch.NewWatchModel(interval)
		p := tea.NewProgram(m, tea.WithAltScreen())

		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	watchCmd.Flags().IntVar(&watchInterval, "interval", 2, "Polling interval in seconds")
	rootCmd.AddCommand(watchCmd)
}
