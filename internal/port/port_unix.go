//go:build !windows

package port

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// ScanPorts returns all listening TCP/UDP ports on Linux/macOS.
func ScanPorts() ([]PortInfo, error) {
	// Try lsof first (works on both Linux and macOS)
	ports, err := scanWithLsof()
	if err == nil {
		return ports, nil
	}

	// Fallback to ss (Linux only)
	return scanWithSS()
}

func scanWithLsof() ([]PortInfo, error) {
	out, err := exec.Command("lsof", "-iTCP", "-sTCP:LISTEN", "-P", "-n", "-F", "pcnT").Output()
	if err != nil {
		return nil, fmt.Errorf("lsof not available: %w", err)
	}

	var ports []PortInfo
	var current PortInfo

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		prefix := line[0]
		value := line[1:]

		switch prefix {
		case 'p': // PID
			if current.PID != 0 && current.Port != 0 {
				ports = append(ports, current)
			}
			current = PortInfo{Proto: "TCP", State: "LISTEN"}
			current.PID, _ = strconv.Atoi(value)
		case 'c': // Command name
			current.ProcessName = value
		case 'n': // Name (address:port)
			current.LocalAddr = value
			lastColon := strings.LastIndex(value, ":")
			if lastColon != -1 {
				current.Port, _ = strconv.Atoi(value[lastColon+1:])
			}
		case 'T': // TCP info
			if strings.HasPrefix(value, "ST=") {
				current.State = value[3:]
			}
		}
	}

	// Don't forget the last entry
	if current.PID != 0 && current.Port != 0 {
		ports = append(ports, current)
	}

	return ports, nil
}

func scanWithSS() ([]PortInfo, error) {
	out, err := exec.Command("ss", "-tlnp").Output()
	if err != nil {
		return nil, fmt.Errorf("ss not available: %w", err)
	}

	var ports []PortInfo
	lines := strings.Split(string(out), "\n")

	for i, line := range lines {
		if i == 0 { // Skip header
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		localAddr := fields[3]
		lastColon := strings.LastIndex(localAddr, ":")
		if lastColon == -1 {
			continue
		}
		portNum, err := strconv.Atoi(localAddr[lastColon+1:])
		if err != nil {
			continue
		}

		// Parse process info from users field
		procName := "unknown"
		pid := 0
		if len(fields) >= 6 {
			usersField := fields[5]
			// Format: users:(("name",pid=123,fd=4))
			if pidIdx := strings.Index(usersField, "pid="); pidIdx != -1 {
				pidStr := usersField[pidIdx+4:]
				if commaIdx := strings.Index(pidStr, ","); commaIdx != -1 {
					pidStr = pidStr[:commaIdx]
				}
				if closeIdx := strings.Index(pidStr, ")"); closeIdx != -1 {
					pidStr = pidStr[:closeIdx]
				}
				pid, _ = strconv.Atoi(pidStr)
			}
			if nameStart := strings.Index(usersField, "((\""); nameStart != -1 {
				nameEnd := strings.Index(usersField[nameStart+3:], "\"")
				if nameEnd != -1 {
					procName = usersField[nameStart+3 : nameStart+3+nameEnd]
				}
			}
		}

		ports = append(ports, PortInfo{
			Proto:       "TCP",
			LocalAddr:   localAddr,
			Port:        portNum,
			PID:         pid,
			ProcessName: procName,
			State:       "LISTEN",
		})
	}

	return ports, nil
}

// killPID kills a process by PID on Unix systems.
func killPID(pid int) error {
	err := syscall.Kill(pid, syscall.SIGKILL)
	if err != nil {
		return fmt.Errorf("failed to kill PID %d: %w", pid, err)
	}
	return nil
}
