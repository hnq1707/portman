//go:build !windows
// +build !windows

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
)

func runSetup() {
	fmt.Println("🚀 Setting up PortMan for macOS/Linux...")

	exePath, err := os.Executable()
	if err != nil {
		color.Red("Error getting executable path: %v", err)
		return
	}

	installDir := "/usr/local/bin"
	targetPath := filepath.Join(installDir, "portman")

	fmt.Printf("📦 To install PortMan, run the following command:\n\n")
	
	// Check if already installed
	if exePath == targetPath {
		color.Green("✅ PortMan is already installed in %s", installDir)
		return
	}

	color.Cyan("sudo cp %s %s && sudo chmod +x %s", exePath, targetPath, targetPath)
	fmt.Println("\nAfter running the command, you can use 'portman' from anywhere.")
}
