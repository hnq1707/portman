//go:build !windows

package port

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// InvestigatePort performs a deep investigation of what's using a port on Unix/macOS.
func InvestigatePort(portNum int) (*ProcessDetail, error) {
	entries, err := FindByPort(portNum)
	if err != nil {
		return nil, fmt.Errorf("failed to scan port %d: %w", portNum, err)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("no process found on port %d — port is free", portNum)
	}

	entry := entries[0]
	detail := &ProcessDetail{
		Port:        portNum,
		PID:         entry.PID,
		ProcessName: entry.ProcessName,
	}

	pidStr := strconv.Itoa(entry.PID)

	// Command line
	if out, err := exec.Command("ps", "-p", pidStr, "-o", "command=").Output(); err == nil {
		detail.CommandLine = strings.TrimSpace(string(out))
	}

	// User
	if out, err := exec.Command("ps", "-p", pidStr, "-o", "user=").Output(); err == nil {
		detail.User = strings.TrimSpace(string(out))
	}

	// Memory (RSS in KB → MB)
	if out, err := exec.Command("ps", "-p", pidStr, "-o", "rss=").Output(); err == nil {
		if rss, err := strconv.Atoi(strings.TrimSpace(string(out))); err == nil {
			detail.MemoryMB = float64(rss) / 1024.0
		}
	}

	// Working directory (Linux)
	if out, err := exec.Command("readlink", fmt.Sprintf("/proc/%d/cwd", entry.PID)).Output(); err == nil {
		detail.WorkingDir = strings.TrimSpace(string(out))
	} else {
		detail.WorkingDir = inferProjectDir(detail.CommandLine, detail.ExecutablePath)
	}

	// Related ports
	allPorts, _ := ScanPorts()
	for _, p := range allPorts {
		if p.PID == entry.PID && p.Port != portNum {
			detail.RelatedPorts = append(detail.RelatedPorts, p)
		}
	}

	// Config files
	if detail.WorkingDir != "" {
		detail.ConfigFiles = scanConfigFiles(detail.WorkingDir, portNum)
	}

	return detail, nil
}
