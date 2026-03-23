package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/nay-kia/portman/internal/ui"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "🖥️  Interactive TUI dashboard for port monitoring",
	Long:  "Launch an interactive terminal dashboard to monitor, filter, and kill ports in realtime.",
	Aliases: []string{"dash", "ui"},
	Run: func(cmd *cobra.Command, args []string) {
		m := ui.NewDashboard()
		p := tea.NewProgram(m, tea.WithAltScreen())

		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
}
