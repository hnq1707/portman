//go:build windows
// +build windows

package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

func runSetup() {
	fmt.Println("🚀 Setting up PortMan for Windows...")

	exePath, err := os.Executable()
	if err != nil {
		color.Red("Error getting executable path: %v", err)
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		color.Red("Error getting home directory: %v", err)
		return
	}

	installDir := filepath.Join(home, ".portman", "bin")
	targetPath := filepath.Join(installDir, "portman.exe")

	// 1. Create directory
	if err := os.MkdirAll(installDir, 0755); err != nil {
		color.Red("Error creating installation directory: %v", err)
		return
	}

	// 2. Copy binary if not already there
	if !strings.EqualFold(exePath, targetPath) {
		fmt.Printf("📦 Copying binary to %s...\n", installDir)
		if err := copyFile(exePath, targetPath); err != nil {
			color.Red("Error copying binary: %v", err)
			fmt.Println("Tip: Try running as Administrator if access is denied.")
			return
		}
	} else {
		fmt.Println("✅ Binary is already in the installation directory.")
	}

	// 3. Add to PATH using PowerShell (Safe method)
	fmt.Println("🔗 Adding to User PATH...")
	
	// PowerShell script to append to User Path if not exists
	psScript := fmt.Sprintf(`
		$Path = [Environment]::GetEnvironmentVariable("Path", "User")
		if ($Path -notlike "*%s*") {
			[Environment]::SetEnvironmentVariable("Path", "$Path;%s", "User")
			Write-Host "SUCCESS"
		} else {
			Write-Host "ALREADY_EXISTS"
		}
	`, installDir, installDir)

	out, err := exec.Command("powershell", "-Command", psScript).Output()
	if err != nil {
		color.Red("Error updating PATH: %v", err)
		return
	}

	result := strings.TrimSpace(string(out))
	if result == "SUCCESS" {
		color.Green("✅ PortMan added to PATH successfully!")
		fmt.Println("\n💡 Please RESTART your terminal (or VS Code) to apply changes.")
	} else if result == "ALREADY_EXISTS" {
		color.Cyan("ℹ️ PortMan is already in your PATH.")
	}

	fmt.Println("\nTry running: portman list")
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
