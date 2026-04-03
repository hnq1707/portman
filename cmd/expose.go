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

	"github.com/hnq1707/portman/internal/tunnel"
)

var (
	exposeProvider string
	exposeLocal    bool
	exposePort     int
	exposeServer   string
)

var exposeCmd = &cobra.Command{
	Use:   "expose <port>",
	Short: "🌐 Expose local port to the network or internet",
	Long: `Expose a local port so other devices can access it.

Modes:
  --local    Expose on LAN (same WiFi/network) — no internet needed
  default    Expose to internet via SSH tunnel (pinggy, serveo, etc.)
  --server   Expose via self-hosted relay server`,
	Example: `  portman expose 3000 --local
  portman expose 3000 --local --port 8080
  portman expose 3000
  portman expose 8080 --provider serveo`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		red := color.New(color.FgRed, color.Bold)

		port, err := strconv.Atoi(args[0])
		if err != nil {
			red.Printf("\n  ✗ Invalid port: %s\n\n", args[0])
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			<-sigCh
			cancel()
		}()

		if exposeLocal {
			// LAN expose
			listenPort := exposePort
			if listenPort == 0 {
				listenPort = port
			}
			if err := tunnel.ExposeLAN(ctx, port, listenPort); err != nil {
				select {
				case <-ctx.Done():
				default:
					red.Printf("\n  ✗ %v\n\n", err)
					os.Exit(1)
				}
			}
		} else {
			// Internet tunnel
			if err := tunnel.Expose(ctx, port, exposeProvider); err != nil {
				select {
				case <-ctx.Done():
				default:
					red.Printf("\n  ✗ %v\n\n", err)
					os.Exit(1)
				}
			}
		}

		fmt.Println()
	},
}

func init() {
	exposeCmd.Flags().BoolVar(&exposeLocal, "local", false, "Expose on LAN only (same network)")
	exposeCmd.Flags().IntVar(&exposePort, "port", 0, "Listen port for LAN expose (default: same as target)")
	exposeCmd.Flags().StringVar(&exposeProvider, "provider", "pinggy", "Tunnel provider (pinggy, localhost.run, serveo)")
	exposeCmd.Flags().StringVar(&exposeServer, "server", "", "Self-hosted relay server URL")
	rootCmd.AddCommand(exposeCmd)
}
