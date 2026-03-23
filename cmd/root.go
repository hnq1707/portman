package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const version = "1.0.0"

var banner = `
  ██████╗  ██████╗ ██████╗ ████████╗███╗   ███╗ █████╗ ███╗   ██╗
  ██╔══██╗██╔═══██╗██╔══██╗╚══██╔══╝████╗ ████║██╔══██╗████╗  ██║
  ██████╔╝██║   ██║██████╔╝   ██║   ██╔████╔██║███████║██╔██╗ ██║
  ██╔═══╝ ██║   ██║██╔══██╗   ██║   ██║╚██╔╝██║██╔══██║██║╚██╗██║
  ██║     ╚██████╔╝██║  ██║   ██║   ██║ ╚═╝ ██║██║  ██║██║ ╚████║
  ╚═╝      ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚═╝     ╚═╝╚═╝  ╚═╝╚═╝  ╚═══╝
`

var rootCmd = &cobra.Command{
	Use:     "portman",
	Short:   "🚀 Local port manager for developers",
	Long:    "PortMan — A blazing fast CLI tool to list, kill, and forward local ports.\nPerfect for microservice development.",
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		cyan := color.New(color.FgCyan, color.Bold)
		cyan.Println(banner)

		yellow := color.New(color.FgYellow)
		yellow.Println("  ⚡ Local Port Manager v" + version)
		fmt.Println()

		white := color.New(color.FgWhite)
		white.Println("  Usage:")
		fmt.Println("    portman list          List all listening ports")
		fmt.Println("    portman kill <port>   Kill process on a port")
		fmt.Println("    portman map <s> <d>   Forward port s → d")
		fmt.Println()
		white.Println("  Run 'portman --help' for more info.")
		fmt.Println()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
