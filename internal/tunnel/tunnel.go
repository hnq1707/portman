package tunnel

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/fatih/color"
)

// Provider defines a tunnel provider.
type Provider struct {
	Name    string
	Host    string
	SSHArgs func(port int) []string
}

// Generic URL regex that catches any https URL
var urlRegex = regexp.MustCompile(`https?://[a-zA-Z0-9\-]+\.[a-zA-Z0-9\.\-]+[/\w]*`)

var providers = map[string]Provider{
	"pinggy": {
		Name: "pinggy.io",
		Host: "a.pinggy.io",
		SSHArgs: func(port int) []string {
			return []string{
				"-p", "443",
				"-o", "StrictHostKeyChecking=no",
				"-o", "ServerAliveInterval=30",
				"-R0:localhost:" + fmt.Sprintf("%d", port),
				"a.pinggy.io",
			}
		},
	},
	"localhost.run": {
		Name: "localhost.run",
		Host: "localhost.run",
		SSHArgs: func(port int) []string {
			return []string{
				"-o", "StrictHostKeyChecking=no",
				"-o", "ServerAliveInterval=30",
				"-R", fmt.Sprintf("80:localhost:%d", port),
				"nokey@localhost.run",
			}
		},
	},
	"serveo": {
		Name: "serveo.net",
		Host: "serveo.net",
		SSHArgs: func(port int) []string {
			return []string{
				"-o", "StrictHostKeyChecking=no",
				"-o", "ServerAliveInterval=30",
				"-R", fmt.Sprintf("80:localhost:%d", port),
				"serveo.net",
			}
		},
	},
}

// Expose creates a public tunnel for a local port.
func Expose(ctx context.Context, port int, providerName string) error {
	provider, ok := providers[providerName]
	if !ok {
		return fmt.Errorf("unknown provider '%s'. Available: pinggy, localhost.run, serveo", providerName)
	}

	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	dim := color.New(color.FgHiBlack)
	yellow := color.New(color.FgYellow)

	// Check SSH
	sshPath := "ssh"
	if runtime.GOOS == "windows" {
		if p, err := exec.LookPath("ssh.exe"); err == nil {
			sshPath = p
		} else if p, err := exec.LookPath("ssh"); err == nil {
			sshPath = p
		} else {
			red.Println("\n  ✗ SSH not found!")
			yellow.Println("  Install OpenSSH:")
			fmt.Println("    Settings → Apps → Optional Features → Add OpenSSH Client")
			fmt.Println()
			return fmt.Errorf("SSH not available")
		}
	}

	// Ensure SSH key exists
	keyPath, err := ensureSSHKey(sshPath)
	if err != nil {
		yellow.Printf("  ⚠ Could not generate SSH key: %v\n", err)
		// Continue anyway, it might still work
	}

	cyan.Printf("\n  🌐 Exposing localhost:%d via %s...\n", port, provider.Name)
	dim.Println("  Connecting to tunnel server...")
	dim.Printf("  Press Ctrl+C to stop\n\n")

	args := provider.SSHArgs(port)
	// Prepend key args if key exists
	if keyPath != "" {
		args = append([]string{"-i", keyPath}, args...)
	}
	cmd := exec.CommandContext(ctx, sshPath, args...)

	// Capture stdout and stderr separately
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start SSH: %w", err)
	}

	urlFound := make(chan string, 1)
	scanDone := make(chan struct{})

	// Scan function for both pipes
	scanPipe := func(reader io.Reader, name string) {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Look for URL
			if url := urlRegex.FindString(line); url != "" {
				select {
				case urlFound <- url:
				default:
				}
			}

			// Debug: show forwarding info
			if strings.Contains(line, "Forwarding") || strings.Contains(line, "forwarding") {
				dim.Printf("  SSH: %s\n", line)
			}
			if strings.Contains(line, "error") || strings.Contains(line, "refused") ||
				strings.Contains(line, "denied") {
				yellow.Printf("  SSH: %s\n", line)
			}
		}
	}

	go func() {
		go scanPipe(stdoutPipe, "stdout")
		go scanPipe(stderrPipe, "stderr")
	}()

	// Wait for URL or process to exit
	go func() {
		cmd.Wait()
		close(scanDone)
	}()

	select {
	case url := <-urlFound:
		fmt.Println()
		green.Println("  ✓ Tunnel is live!")
		fmt.Println()
		cyan.Printf("  🔗 Public URL: %s\n", url)
		fmt.Println()
		dim.Printf("  → localhost:%d ←→ %s\n", port, url)
		dim.Println("  Press Ctrl+C to close tunnel")
		fmt.Println()

		// Wait for shutdown
		select {
		case <-ctx.Done():
			yellow.Println("\n  ⏹ Tunnel closed.")
			fmt.Println()
			return nil
		case <-scanDone:
			return nil
		}

	case <-scanDone:
		// Process exited without URL
		select {
		case <-ctx.Done():
			yellow.Println("\n  ⏹ Tunnel closed.")
			fmt.Println()
			return nil
		default:
			red.Println("\n  ✗ Tunnel failed. SSH exited without providing a URL.")
			yellow.Println("  Try a different provider: portman expose <port> --provider serveo")
			fmt.Println()
			return fmt.Errorf("tunnel failed")
		}

	case <-ctx.Done():
		yellow.Println("\n  ⏹ Tunnel closed.")
		fmt.Println()
		return nil
	}
}

// ListProviders returns available provider names.
func ListProviders() []string {
	var names []string
	for name := range providers {
		names = append(names, name)
	}
	return names
}

// ensureSSHKey checks if an SSH key exists and generates one if not.
func ensureSSHKey(sshPath string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	keyPath := filepath.Join(home, ".ssh", "id_rsa")

	// Check if key already exists
	if _, err := os.Stat(keyPath); err == nil {
		return keyPath, nil
	}

	// Check for ed25519 key
	ed25519Path := filepath.Join(home, ".ssh", "id_ed25519")
	if _, err := os.Stat(ed25519Path); err == nil {
		return ed25519Path, nil
	}

	// Generate new SSH key
	dim := color.New(color.FgHiBlack)
	dim.Println("  🔑 Generating SSH key...")

	sshDir := filepath.Join(home, ".ssh")
	os.MkdirAll(sshDir, 0700)

	keygenPath := "ssh-keygen"
	if runtime.GOOS == "windows" {
		if p, e := exec.LookPath("ssh-keygen.exe"); e == nil {
			keygenPath = p
		}
	}

	cmd := exec.Command(keygenPath, "-t", "rsa", "-b", "4096", "-f", keyPath, "-N", "")
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ssh-keygen failed: %w", err)
	}

	dim.Println("  ✓ SSH key generated")
	return keyPath, nil
}
