package cmd

import (
	"fmt"
	"os"

	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const version = "1.2.0"

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
		dim := color.New(color.FgHiBlack)

		white.Println("  Core:")
		fmt.Println("    portman list              List all listening ports")
		fmt.Println("    portman kill <port>        Kill process on a port")
		fmt.Println("    portman why <port>         Deep investigate a port")
		fmt.Println()

		white.Println("  Monitor:")
		fmt.Println("    portman watch             Real-time port activity timeline")
		fmt.Println("    portman dashboard         Interactive TUI dashboard")
		fmt.Println("    portman doctor            Diagnose port health issues")
		fmt.Println()

		white.Println("  Network:")
		fmt.Println("    portman map <s> <d>       Forward port s → d")
		fmt.Println("    portman expose <port>     Expose port to internet/LAN")
		fmt.Println("    portman docker            Show Docker port bindings")
		fmt.Println()

		white.Println("  Config:")
		fmt.Println("    portman profile           Manage port profiles")
		fmt.Println()

		dim.Println("  Run 'portman <command> --help' for more info.")
		fmt.Println()
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip path check for setup command itself
		if cmd.Name() == "setup" || cmd.Name() == "help" || cmd.Name() == "version" {
			return
		}
		checkPath()
	},
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "🚀 Install PortMan to your system PATH",
	Run: func(cmd *cobra.Command, args []string) {
		runSetup()
	},
}

func checkPath() {
	exePath, err := os.Executable()
	if err != nil {
		return
	}

	exeDir := filepath.Dir(exePath)
	pathEnv := os.Getenv("PATH")

	// Check if exeDir is in PATH
	inPath := false
	for _, p := range filepath.SplitList(pathEnv) {
		if strings.EqualFold(p, exeDir) {
			inPath = true
			break
		}
	}

	if !inPath {
		yellow := color.New(color.FgYellow, color.Bold)
		fmt.Println()
		yellow.Println("  ⚠️  PortMan is not in your system PATH.")
		fmt.Println("     To run 'portman' from anywhere, please run:")
		color.Cyan("     portman setup")
		fmt.Println()
	}
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
