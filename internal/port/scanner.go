package port

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// PortInfo holds information about a port binding.
type PortInfo struct {
	Proto       string `json:"proto"`
	LocalAddr   string `json:"local_addr"`
	Port        int    `json:"port"`
	PID         int    `json:"pid"`
	ProcessName string `json:"process_name"`
	State       string `json:"state"`
}

// ScanPorts returns all listening TCP/UDP ports on the system.
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

		// Parse port from local address
		lastColon := strings.LastIndex(localAddr, ":")
		if lastColon == -1 {
			continue
		}
		portStr := localAddr[lastColon+1:]
		portNum, err := strconv.Atoi(portStr)
		if err != nil {
			continue
		}

		// State and PID
		var state string
		var pidStr string

		if proto == "TCP" {
			if len(fields) < 5 {
				continue
			}
			state = fields[3]
			pidStr = fields[4]
		} else {
			// UDP has no state column
			state = "*"
			pidStr = fields[3]
		}

		// Only show LISTENING ports (and all UDP)
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

// FindByPort returns port info entries matching the given port number.
func FindByPort(port int) ([]PortInfo, error) {
	all, err := ScanPorts()
	if err != nil {
		return nil, err
	}

	var matched []PortInfo
	for _, p := range all {
		if p.Port == port {
			matched = append(matched, p)
		}
	}
	return matched, nil
}

// buildPIDNameMap uses tasklist to map PIDs to process names.
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
		// Format: "name.exe","PID","Session Name","Session#","Mem Usage"
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
