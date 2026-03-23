package port

import (
	"fmt"
	"os/exec"
	"strconv"
)

// KillResult holds the result of a kill operation.
type KillResult struct {
	Port        int
	PID         int
	ProcessName string
	Success     bool
	Error       string
}

// KillByPort finds the process using the given port and kills it.
func KillByPort(port int) []KillResult {
	entries, err := FindByPort(port)
	if err != nil {
		return []KillResult{{
			Port:    port,
			Success: false,
			Error:   fmt.Sprintf("failed to scan ports: %v", err),
		}}
	}

	if len(entries) == 0 {
		return []KillResult{{
			Port:    port,
			Success: false,
			Error:   fmt.Sprintf("no process found on port %d", port),
		}}
	}

	// Deduplicate PIDs (same PID can appear on multiple addresses)
	seen := make(map[int]bool)
	var results []KillResult

	for _, entry := range entries {
		if seen[entry.PID] {
			continue
		}
		seen[entry.PID] = true

		err := killPID(entry.PID)
		result := KillResult{
			Port:        port,
			PID:         entry.PID,
			ProcessName: entry.ProcessName,
		}
		if err != nil {
			result.Success = false
			result.Error = err.Error()
		} else {
			result.Success = true
		}
		results = append(results, result)
	}

	return results
}

// killPID kills a process by its PID using taskkill.
func killPID(pid int) error {
	cmd := exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/F")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, string(output))
	}
	return nil
}
