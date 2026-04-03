package port

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ProcessDetail holds detailed investigation results for a port.
type ProcessDetail struct {
	Port           int
	PID            int
	ProcessName    string
	CommandLine    string
	ExecutablePath string
	WorkingDir     string
	User           string
	MemoryMB       float64
	StartTime      time.Time
	ParentPID      int
	ProcessTree    []TreeNode
	RelatedPorts   []PortInfo
	ConfigFiles    []ConfigMatch
}

// TreeNode represents a node in the process tree.
type TreeNode struct {
	PID       int
	Name      string
	Depth     int
	IsCurrent bool
}

// ConfigMatch represents a config file that references the port.
type ConfigMatch struct {
	FilePath string
	Line     string
	LineNum  int
}

// FormatDuration formats a duration into a human-readable string.
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh %dm ago", int(d.Hours()), int(d.Minutes())%60)
	}
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%dd %dh ago", days, hours)
}

// inferProjectDir tries to find the project directory from command line args.
func inferProjectDir(cmdLine, exePath string) string {
	if cmdLine != "" {
		parts := splitCommandLine(cmdLine)
		// Check arguments (skip the executable itself)
		for i := len(parts) - 1; i >= 1; i-- {
			part := strings.Trim(parts[i], "\"'")
			if part == "" || strings.HasPrefix(part, "-") {
				continue
			}

			// Check if it's an absolute path to a file
			if filepath.IsAbs(part) {
				dir := filepath.Dir(part)
				if info, err := os.Stat(dir); err == nil && info.IsDir() {
					return dir
				}
			}
		}
	}

	if exePath != "" {
		return filepath.Dir(exePath)
	}
	return ""
}

func splitCommandLine(cmdLine string) []string {
	var parts []string
	var current strings.Builder
	inQuote := false
	quoteChar := byte(0)

	for i := 0; i < len(cmdLine); i++ {
		c := cmdLine[i]
		if inQuote {
			if c == quoteChar {
				inQuote = false
			}
			current.WriteByte(c)
		} else if c == '"' || c == '\'' {
			inQuote = true
			quoteChar = c
			current.WriteByte(c)
		} else if c == ' ' || c == '\t' {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else {
			current.WriteByte(c)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

// scanConfigFiles searches for config files that reference the given port.
func scanConfigFiles(dir string, portNum int) []ConfigMatch {
	configNames := []string{
		"package.json",
		".env", ".env.local", ".env.development", ".env.production",
		"docker-compose.yml", "docker-compose.yaml",
		".portman.yml",
		"application.properties", "application.yml", "application.yaml",
	}

	portStr := strconv.Itoa(portNum)
	var matches []ConfigMatch

	// Check current dir and up to 2 parent levels
	searchDirs := []string{dir}
	current := dir
	for i := 0; i < 2; i++ {
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		searchDirs = append(searchDirs, parent)
		current = parent
	}

	seen := make(map[string]bool)
	for _, d := range searchDirs {
		for _, name := range configNames {
			path := filepath.Join(d, name)
			if seen[path] {
				continue
			}
			seen[path] = true

			f, err := os.Open(path)
			if err != nil {
				continue
			}

			scanner := bufio.NewScanner(f)
			lineNum := 0
			for scanner.Scan() {
				lineNum++
				line := scanner.Text()
				if strings.Contains(line, portStr) {
					matches = append(matches, ConfigMatch{
						FilePath: path,
						Line:     strings.TrimSpace(line),
						LineNum:  lineNum,
					})
				}
			}
			f.Close()
		}
	}

	return matches
}
