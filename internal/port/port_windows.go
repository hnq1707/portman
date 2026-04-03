//go:build windows

package port

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// ScanPorts returns all listening TCP/UDP ports on Windows.
func ScanPorts() ([]PortInfo, error) {
	out, err := exec.Command("netstat", "-ano").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run netstat: %w", err)
	}

	pidNameMap, _ := buildPIDNameMap()

	var ports []PortInfo
	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		proto := strings.ToUpper(fields[0])
		if proto != "TCP" && proto != "UDP" {
			continue
		}

		localAddr := fields[1]

		lastColon := strings.LastIndex(localAddr, ":")
		if lastColon == -1 {
			continue
		}
		portStr := localAddr[lastColon+1:]
		portNum, err := strconv.Atoi(portStr)
		if err != nil {
			continue
		}

		var state string
		var pidStr string

		if proto == "TCP" {
			if len(fields) < 5 {
				continue
			}
			state = fields[3]
			pidStr = fields[4]
		} else {
			state = "*"
			pidStr = fields[3]
		}

		if proto == "TCP" && state != "LISTENING" {
			continue
		}

		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}

		procName := "unknown"
		if name, ok := pidNameMap[pid]; ok {
			procName = name
		}

		ports = append(ports, PortInfo{
			Proto:       proto,
			LocalAddr:   localAddr,
			Port:        portNum,
			PID:         pid,
			ProcessName: procName,
			State:       state,
		})
	}

	return ports, nil
}

func buildPIDNameMap() (map[int]string, error) {
	out, err := exec.Command("tasklist", "/FO", "CSV", "/NH").Output()
	if err != nil {
		return nil, err
	}

	m := make(map[int]string)
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\",\"")
		if len(parts) < 2 {
			continue
		}
		name := strings.Trim(parts[0], "\"")
		pidStr := strings.Trim(parts[1], "\"")
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}
		m[pid] = name
	}
	return m, nil
}

// killPID kills a process by PID on Windows.
func killPID(pid int) error {
	cmd := exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/F")
	output, err := cmd.CombinedOutput()
	if err != nil {
		outStr := string(output)
		if strings.Contains(outStr, "Access is denied") {
			return fmt.Errorf("Access Denied: Administrator privileges required")
		}
		if strings.Contains(outStr, "not found") {
			return fmt.Errorf("Process ID %d not found", pid)
		}
		return fmt.Errorf("%s: %s", err, outStr)
	}
	return nil
}
