package tunnel

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
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
	URLRe   *regexp.Regexp
}

var providers = map[string]Provider{
	"localhost.run": {
		Name: "localhost.run",
		Host: "localhost.run",
		SSHArgs: func(port int) []string {
			return []string{
				"-o", "StrictHostKeyChecking=no",
				"-o", "ServerAliveInterval=30",
				"-R", fmt.Sprintf("80:localhost:%d", port),
				"localhost.run",
			}
		},
		URLRe: regexp.MustCompile(`https://[a-zA-Z0-9\-]+\.lhr\.life`),
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
		URLRe: regexp.MustCompile(`https://[a-zA-Z0-9\-]+\.serveo\.net`),
	},
}

// Expose creates a public tunnel for a local port.
func Expose(ctx context.Context, port int, providerName string) error {
	provider, ok := providers[providerName]
	if !ok {
		return fmt.Errorf("unknown provider '%s'. Available: localhost.run, serveo", providerName)
	}

	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	dim := color.New(color.FgHiBlack)
	yellow := color.New(color.FgYellow)

	// Check SSH
	sshPath := "ssh"
	if runtime.GOOS == "windows" {
		// Try to find ssh.exe
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

	cyan.Printf("\n  🌐 Exposing localhost:%d via %s...\n", port, provider.Name)
	dim.Println("  Connecting to tunnel server...")
	dim.Println("  Press Ctrl+C to stop\n")

	args := provider.SSHArgs(port)
	cmd := exec.CommandContext(ctx, sshPath, args...)

	// Capture both stdout and stderr (URL can appear in either)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	cmd.Stderr = cmd.Stdout // merge stderr into stdout

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start SSH: %w", err)
	}

	// Read output looking for the URL
	scanner := bufio.NewScanner(stdout)
	urlFound := false

	go func() {
		for scanner.Scan() {
			line := scanner.Text()

			if !urlFound {
				if url := provider.URLRe.FindString(line); url != "" {
					urlFound = true
					fmt.Println()
					green.Println("  ✓ Tunnel is live!")
					fmt.Println()
					cyan.Printf("  🔗 Public URL: %s\n", url)
					fmt.Println()
					dim.Printf("  → localhost:%d ←→ %s\n", port, url)
					dim.Println("  Press Ctrl+C to close tunnel")
					fmt.Println()
				}
			}

			// Print SSH output for debugging
			if strings.Contains(line, "Warning") || strings.Contains(line, "Error") ||
				strings.Contains(line, "error") || strings.Contains(line, "refused") {
				yellow.Printf("  SSH: %s\n", line)
			}
		}
	}()

	// Wait for process to exit or context cancellation
	err = cmd.Wait()

	select {
	case <-ctx.Done():
		yellow.Println("\n  ⏹ Tunnel closed.")
		fmt.Println()
		return nil
	default:
		if err != nil && !urlFound {
			red.Printf("\n  ✗ SSH tunnel failed: %v\n", err)
			yellow.Println("  Make sure SSH is installed and port is valid.")
			fmt.Println()
		}
		return err
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
