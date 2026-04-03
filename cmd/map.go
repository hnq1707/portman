package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/hnq1707/portman/internal/port"
)

var mapHost string

var mapCmd = &cobra.Command{
	Use:   "map <src-port> <dst-port>",
	Short: "🔗 Forward traffic from one port to another",
	Long:  "Create a TCP proxy that forwards all traffic from src-port to dst-port.\nUseful for redirecting traffic between services during development.",
	Example: `  portman map 3000 8080          # Forward :3000 → :8080
  portman map 9000 3306          # Forward :9000 → MySQL
  portman map 80 3000 --host 0.0.0.0   # Expose to network`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		red := color.New(color.FgRed, color.Bold)

		srcPort, err := strconv.Atoi(args[0])
		if err != nil {
			red.Fprintf(cmd.ErrOrStderr(), "\n  ✗ Invalid source port: %s\n\n", args[0])
			return
		}

		dstPort, err := strconv.Atoi(args[1])
		if err != nil {
			red.Fprintf(cmd.ErrOrStderr(), "\n  ✗ Invalid destination port: %s\n\n", args[1])
			return
		}

		if srcPort == dstPort {
			red.Fprintf(cmd.ErrOrStderr(), "\n  ✗ Source and destination ports must be different\n\n")
			return
		}

		// Graceful shutdown on Ctrl+C
		ctx, cancel := context.WithCancel(context.Background())
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			<-sigCh
			yellow := color.New(color.FgYellow)
			yellow.Printf("\n  ⏹ Shutting down port mapping...\n\n")
			cancel()
		}()

		mapper := port.NewMapper(srcPort, dstPort, mapHost)
		if err := mapper.Start(ctx); err != nil {
			red.Fprintf(cmd.ErrOrStderr(), "\n  ✗ Mapping error: %v\n\n", err)
			os.Exit(1)
		}

		green := color.New(color.FgGreen)
		green.Println("  ✓ Port mapping stopped cleanly.")
		fmt.Println()
	},
}

func init() {
	mapCmd.Flags().StringVar(&mapHost, "host", "127.0.0.1", "Host address to listen on")
	rootCmd.AddCommand(mapCmd)
}
