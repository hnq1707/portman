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

	"github.com/nay-kia/portman/internal/tunnel"
)

var exposeProvider string

var exposeCmd = &cobra.Command{
	Use:   "expose <port>",
	Short: "🌐 Expose local port to the internet",
	Long:  "Create a public tunnel to expose a local port to the internet via SSH.\nNo signup required — uses free tunnel providers.",
	Example: `  portman expose 3000
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

		if err := tunnel.Expose(ctx, port, exposeProvider); err != nil {
			// Only show error if it wasn't a clean shutdown
			select {
			case <-ctx.Done():
			default:
				red.Printf("\n  ✗ %v\n\n", err)
				os.Exit(1)
			}
		}

		fmt.Println()
	},
}

func init() {
	exposeCmd.Flags().StringVar(&exposeProvider, "provider", "localhost.run", "Tunnel provider (localhost.run, serveo)")
	rootCmd.AddCommand(exposeCmd)
}
